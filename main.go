package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
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

var (
	StreamPrefix []byte = []byte("data: ")
	MessageStart []byte = []byte("event: message_start")
	MessageStop  []byte = []byte("event: message_stop")
)

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
			log.Println("Refreshed access token")
			foxy.auth = auth
			foxy.expires_at = time.Now().Add(time.Duration(auth.Expires_in) * time.Second)
		}
		foxy.mu.Unlock()
		log.Printf("%s %s", r.Method, r.URL.Path)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("read body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", foxy.auth.Access_token))

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Do(req)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		return
	}
	defer resp.Body.Close()
	maps.Copy(w.Header(), resp.Header)

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Bytes()
		w.Write(line)
		w.Write([]byte("\n"))
		flusher.Flush()

		if bytes.HasPrefix(line, MessageStop) {
			return
		}
	}
}

func set_required_headers(r *http.Request) {
	for k, v := range RequiredHeaders {
		r.Header.Set(k, v)
	}
}
func filter_excluded_headers(r *http.Request) {
	for _, k := range ExcludedHeaders {
		r.Header.Del(k)
	}
}
