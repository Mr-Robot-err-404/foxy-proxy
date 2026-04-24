package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/cli/browser"
)

// Note: auth.go log.Fatal calls are intentional — auth runs before the TUI/log file is ready

//go:embed html/success.html
var successHTML []byte

//go:embed html/error.html
var errorHTML []byte

const (
	RedirectUri     string = "http://localhost:58388/callback"
	ClientID        string = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	Scope           string = "org:create_api_key user:profile user:inference user:sessions:claude_code user:mcp_servers user:file_upload"
	AuthStart       string = "https://claude.ai/oauth/authorize"
	TokenExchange   string = "https://platform.claude.com/v1/oauth/token"
	ChallengeMethod string = "S256"
	AuthPort        string = ":58388"
)

type Exchange struct {
	Grant    string `json:"grant_type"`
	Code     string `json:"code"`
	Redirect string `json:"redirect_uri"`
	ClientID string `json:"client_id"`
	Verifier string `json:"code_verifier"`
	State    string `json:"state"`
}
type ExchangeResponse struct {
	Access_token  string `json:"access_token"`
	Refresh_token string `json:"refresh_token"`
	Expires_in    int    `json:"expires_in"`
}
type TokenChannel struct {
	Resp ExchangeResponse
	Err  error
}

type AuthCfg struct {
	Verifier string
	State    string
}

func (foxy *Foxy) Auth() {
	ch := make(chan TokenChannel)

	cfg, err := setupAuth()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", HandleAuthCallback(&cfg, ch))
	srv := &http.Server{Handler: mux, Addr: AuthPort}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	err = OpenAuthStart(&cfg)

	if err != nil {
		log.Fatal(err)
	}
	token := <-ch

	if token.Err != nil {
		log.Fatal(token.Err)
	}
	err = foxy.saveAuth(token.Resp)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("oauth saved to file")
}

func ExchangeCode(cfg *AuthCfg, code string) (ExchangeResponse, error) {
	var token ExchangeResponse

	payload := Exchange{
		Grant:    "authorization_code",
		Code:     code,
		Redirect: RedirectUri,
		ClientID: ClientID,
		Verifier: cfg.Verifier,
		State:    cfg.State,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return token, err
	}
	req, err := http.NewRequest(http.MethodPost, TokenExchange, bytes.NewReader(body))

	if err != nil {
		return token, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return token, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return token, fmt.Errorf("%s: %s", resp.Status, body)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return token, err
	}
	return token, nil
}

func HandleAuthCallback(cfg *AuthCfg, ch chan<- TokenChannel) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		if len(code) == 0 {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		token, err := ExchangeCode(cfg, code)

		if err != nil {
			ch <- TokenChannel{Err: err}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errorHTML)
			return
		}
		ch <- TokenChannel{Resp: token}
		w.WriteHeader(http.StatusOK)
		w.Write(successHTML)
	}
}
func OpenAuthStart(cfg *AuthCfg) error {
	u, err := url.Parse(AuthStart)
	if err != nil {
		return err
	}
	query := u.Query()
	query.Set("code", "true")
	query.Set("response_type", "code")
	query.Set("redirect_uri", RedirectUri)
	query.Set("scope", Scope)
	query.Set("code_challenge", HashCodeVerifier(cfg.Verifier))
	query.Set("code_challenge_method", ChallengeMethod)
	query.Set("state", cfg.State)
	query.Set("client_id", ClientID)
	u.RawQuery = query.Encode()

	if err := browser.OpenURL(u.String()); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}
	return nil
}

func setupAuth() (AuthCfg, error) {
	verifier, err := GenerateVerifier()

	if err != nil {
		return AuthCfg{}, err
	}
	state, err := GenerateVerifier()
	if err != nil {
		return AuthCfg{}, err
	}
	return AuthCfg{Verifier: verifier, State: state}, nil
}
