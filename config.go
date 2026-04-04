package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Foxy struct {
	auth ExchangeResponse
	root string
	port string
}

const (
	AuthFile string = "auth.json"
	FoxyPath string = "/.config/foxy/"
)

func initFoxy() (Foxy, error) {
	root, err := rootPath()

	if err != nil {
		return Foxy{}, err
	}
	cfg := Foxy{root: root, port: Port}
	auth, err := cfg.readAuth()

	if err != nil {
		return cfg, err
	}
	cfg.auth = auth
	return cfg, nil
}

func rootPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home + FoxyPath, nil
}

func (foxy *Foxy) saveAuth(auth ExchangeResponse) error {
	b, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	path := foxy.root + AuthFile
	return os.WriteFile(path, b, 0644)
}

func (foxy *Foxy) readAuth() (ExchangeResponse, error) {
	var auth ExchangeResponse

	path := foxy.root + AuthFile
	b, err := os.ReadFile(path)

	if err != nil {
		return auth, err
	}
	err = json.Unmarshal(b, &auth)

	if err != nil {
		return auth, err
	}
	return auth, nil
}

func (foxy *Foxy) server_url() string {
	return fmt.Sprintf("http://localhost%s", foxy.port)
}
