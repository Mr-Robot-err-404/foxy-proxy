package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequiredHeadersPreserveCallerBetas(t *testing.T) {
	r := httptest.NewRequest("POST", "/v1/messages", nil)
	r.Header.Add("anthropic-beta", "thinking-binding-controls-2026-08-01, oauth-2025-04-20")
	r.Header.Add("anthropic-beta", "custom-beta, thinking-binding-controls-2026-08-01, ")
	r.Header.Set("User-Agent", "caller")
	set_required_headers(r)

	counts := make(map[string]int)
	for beta := range strings.SplitSeq(r.Header.Get("anthropic-beta"), ",") {
		counts[beta]++
	}
	for _, beta := range append(strings.Split(RequiredHeaders["anthropic-beta"], ","), "thinking-binding-controls-2026-08-01", "custom-beta") {
		if counts[beta] != 1 {
			t.Errorf("beta %q occurs %d times, want 1", beta, counts[beta])
		}
	}
	if counts[""] != 0 {
		t.Error("empty beta retained")
	}
	if r.Header.Get("User-Agent") != RequiredHeaders["User-Agent"] {
		t.Error("required user agent not applied")
	}
}

func TestRequiredHeadersWithoutCallerBetas(t *testing.T) {
	r := httptest.NewRequest("POST", "/v1/messages", nil)
	set_required_headers(r)
	if got := r.Header.Get("anthropic-beta"); got != RequiredHeaders["anthropic-beta"] {
		t.Fatalf("got betas %q, want %q", got, RequiredHeaders["anthropic-beta"])
	}
}
