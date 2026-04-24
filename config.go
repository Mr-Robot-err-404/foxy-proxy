package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type Foxy struct {
	auth       ExchangeResponse
	root       string
	port       string
	expires_at time.Time
	mu         *sync.Mutex
}

const (
	AuthFile   string = "auth.json"
	ExpiryFile string = "expiry"
	LogFile    string = "foxy.log"
	FoxyPath   string = "/.config/foxy/"
)

func init_foxy() (Foxy, error) {
	root, err := root_path()

	if err != nil {
		return Foxy{}, err
	}
	return Foxy{root: root, port: Port, mu: new(sync.Mutex)}, nil
}

func (foxy *Foxy) complete_setup() error {
	if err := foxy.init_logger(); err != nil {
		return err
	}
	auth, err := foxy.read_auth()

	if err != nil {
		return err
	}
	expiry, err := foxy.read_expiry()

	if err != nil {
		return err
	}
	foxy.auth = auth
	foxy.expires_at = time.Unix(expiry, 0)
	return nil
}

func (foxy *Foxy) init_logger() error {
	path := foxy.root + LogFile
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	return nil
}

func root_path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home + FoxyPath, nil
}

func (foxy *Foxy) save_auth(auth ExchangeResponse) error {
	b, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	if err := os.WriteFile(foxy.root+AuthFile, b, 0600); err != nil {
		return err
	}
	expiry := time.Now().Unix() + int64(auth.Expires_in)
	return os.WriteFile(foxy.root+ExpiryFile, []byte(strconv.FormatInt(expiry, 10)), 0600)
}

func (foxy *Foxy) read_auth() (ExchangeResponse, error) {
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

func (foxy *Foxy) should_refresh() bool {
	return time.Now().After(foxy.expires_at.Add(-5 * time.Minute))
}

func (foxy *Foxy) read_expiry() (int64, error) {
	b, err := os.ReadFile(foxy.root + ExpiryFile)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(b), 10, 64)
}

func (foxy *Foxy) server_url() string {
	return fmt.Sprintf("http://localhost%s", foxy.port)
}
