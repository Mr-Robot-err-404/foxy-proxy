package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type ServerConfig struct {
	Bearer string
}
type APIFunc func(w http.ResponseWriter, r *http.Request)

const Port string = ":6942"
const TargetServer string = "https://api.anthropic.com/v1/messages?beta=true"

func main() {
	dev := flag.Bool("dev", false, "dev mode")
	flag.Parse()

	if *dev {
		tinker()
		return
	}
	bearer, err := os.ReadFile("bearer")
	if err != nil {
		panic(err)
	}
	cfg := &ServerConfig{Bearer: strings.TrimSpace(string(bearer))}

	mux := http.NewServeMux()
	mux.HandleFunc("/", messageHandler(cfg))

	fmt.Printf("listening on port %s\n", Port)

	err = http.ListenAndServe(Port, mux)
	if err != nil {
		panic(err)
	}
}

func messageHandler(cfg *ServerConfig) APIFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("=== INCOMING REQUEST ===")
		fmt.Println(r.Method, r.URL.Path)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Body.Close()
		payload, err := sanitize_payload(body)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		set_required_headers(r)
		filter_excluded_headers(r)
		forward(cfg, w, r, payload)
	}
}

func forward(cfg *ServerConfig, w http.ResponseWriter, r *http.Request, payload []byte) {
	req, err := http.NewRequestWithContext(r.Context(), "POST", TargetServer, bytes.NewReader(payload))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.Header = r.Header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.Bearer))

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
	defer resp.Body.Close()

	fmt.Println(resp.Status)
	w.WriteHeader(resp.StatusCode)
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
