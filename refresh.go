package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RefreshResponse struct {
	Token_type    string `json:"token_type"`
	Access_token  string `json:"access_token"`
	Expires_in    int    `json:"expires_in"`
	Refresh_token string `json:"refresh_token"`
	Scope         string `json:"scope"`
}

type RefreshPayload struct {
	Grant_type    string `json:"grant_type"`
	Refresh_token string `json:"refresh_token"`
	ClientID      string `json:"client_id"`
	Scope         string `json:"scope"`
}

var RefreshHeaders = map[string]string{
	"Content-Type": "application/json",
	"User-agent":   "axios/1.13.6",
	"Accept":       "application/json",
	"Host":         "platform.claude.com",
}

const RefreshURL string = "https://platform.claude.com/v1/oauth/token"
const RefreshScope string = "user:profile user:inference user:sessions:claude_code user:mcp_servers user:file_upload"

func exchange_refresh_token(refresh_token string) (RefreshResponse, error) {
	var result RefreshResponse

	payload := RefreshPayload{
		Grant_type:    "refresh_token",
		Refresh_token: refresh_token,
		ClientID:      ClientID,
		Scope:         RefreshScope,
	}
	body, err := json.Marshal(payload)

	if err != nil {
		return result, err
	}
	req, err := http.NewRequest(http.MethodPost, RefreshURL, bytes.NewReader(body))

	if err != nil {
		return result, err
	}
	for k, v := range RefreshHeaders {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("%s: %s", resp.Status, b)
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}
