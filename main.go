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
)

type ServerConfig struct {
	Bearer string
}
type APIFunc func(w http.ResponseWriter, r *http.Request)

const Port string = ":6942"
const TargetServer string = "https://api.anthropic.com/v1/messages?beta=true"

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
	foxy, err := initFoxy()

	if err != nil {
		log.Fatal(err)
	}
	if *auth {
		foxy.Auth()
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", message_handler(&foxy))

	fmt.Printf("listening on port %s\n", foxy.port)

	err = http.ListenAndServe(foxy.port, mux)
	if err != nil {
		panic(err)
	}
}

func message_handler(foxy *Foxy) APIFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
