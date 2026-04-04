package main

import (
	"fmt"
	"net/http"
)

func (foxy *Foxy) is_server_live() bool {
	url := fmt.Sprintf("%s/foxy/health", foxy.server_url())
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func (foxy *Foxy) serve() {
	if foxy.is_server_live() {
		fmt.Printf("server already running on port %s\n", foxy.port)
		return
	}
	foxy.run_server()
}
