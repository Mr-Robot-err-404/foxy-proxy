package main

import (
	"encoding/json"
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

func (cfg *Foxy) saveAuth(auth ExchangeResponse) error {
	b, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	path := cfg.root + AuthFile
	return os.WriteFile(path, b, 0644)
}

func (cfg *Foxy) readAuth() (ExchangeResponse, error) {
	var auth ExchangeResponse

	path := cfg.root + AuthFile
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
