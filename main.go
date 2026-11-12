package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"strings"
	"time"
)

type ServerConfig struct {
	Bearer string
}
type APIFunc func(w http.ResponseWriter, r *http.Request)

const (
	Port         string = ":6942"
	TargetServer string = "https://api.anthropic.com/v1/messages?beta=true"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	},
}

func main() {
	dev := flag.Bool("dev", false, "dev mode")
	auth := flag.Bool("auth", false, "auth flow")
	flag.Parse()

	if *dev {
		tinker()
		return
	}
	foxy, err := init_foxy()

	if err != nil {
		log.Fatal(err)
	}
	if *auth {
		foxy.Auth()
		return
	}
	err = foxy.complete_setup()

	if err != nil {
		log.Fatal(err)
	}
	go foxy.serve()
	if err := run_tui(foxy.port[1:], foxy.root+LogFile); err != nil {
		log.Fatal(err)
	}
}

func (foxy *Foxy) run_server() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages", message_handler(foxy))
	mux.HandleFunc("/foxy/health", health_check())

	err := http.ListenAndServe(foxy.port, mux)
	if err != nil {
		panic(err)
	}
}

func health_check() APIFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foxy is foxing"))
	}
}

func message_handler(foxy *Foxy) APIFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		foxy.mu.Lock()

		if foxy.should_refresh() {
			auth, err := foxy.exchange_refresh_token()

			if err != nil {
				log.Printf("failed to refresh token: %s", err)
				foxy.mu.Unlock()
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := foxy.save_auth(auth); err != nil {
				log.Printf("failed to save auth: %s", err)
			}
			log.Printf("Refreshed access token, expires_in=%d expires_at=%s", auth.ExpiresIn, time.Now().Add(time.Duration(auth.ExpiresIn)*time.Second).Format(time.RFC3339))
			foxy.auth = auth
			foxy.expires_at = time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)
		}
		foxy.mu.Unlock()
		log.Printf("%s %s", r.Method, r.URL.Path)

		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("read body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		payload, err := sanitize_payload(body)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		set_required_headers(r)
		filter_excluded_headers(r)
		forward(foxy, w, r, payload)
	}
}

func forward(foxy *Foxy, w http.ResponseWriter, r *http.Request, payload []byte) {
	req, err := http.NewRequestWithContext(r.Context(), "POST", TargetServer, bytes.NewReader(payload))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.Header = r.Header.Clone()
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", foxy.auth.AccessToken))

	resp, err := httpClient.Do(req)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		maps.Copy(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}
	maps.Copy(w.Header(), resp.Header)

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := relay_stream(w, resp.Body, flusher.Flush); err != nil {
		log.Printf("stream relay failed: %v", err)
	}
}

func set_required_headers(r *http.Request) {
	for k, v := range RequiredHeaders {
		if strings.EqualFold(k, "anthropic-beta") {
			values := append(r.Header.Values(k), v)
			seen := make(map[string]bool)
			var betas []string
			for _, value := range values {
				for beta := range strings.SplitSeq(value, ",") {
					beta = strings.TrimSpace(beta)
					if beta != "" && !seen[beta] {
						seen[beta] = true
						betas = append(betas, beta)
					}
				}
			}
			r.Header.Set(k, strings.Join(betas, ","))
			continue
		}
		r.Header.Set(k, v)
	}
}
func filter_excluded_headers(r *http.Request) {
	for _, k := range ExcludedHeaders {
		r.Header.Del(k)
	}
}
