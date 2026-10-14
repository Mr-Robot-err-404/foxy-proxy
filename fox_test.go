package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizePayloadDropsEmptySystemTextBlocks(t *testing.T) {
	input := []byte(`{"system":[{"type":"text","text":"opencode"},{"type":"text","text":"<directories>\n</directories>"},{"type":"text","text":"Keep this instruction\nopencode"}],"messages":[{"role":"user","content":"hi"}]}`)
	output, err := sanitize_payload(input)
	if err != nil {
		t.Fatal(err)
	}

	var payload struct {
		System []SystemItem `json:"system"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.System) != 2 {
		t.Fatalf("got %d system blocks, want fingerprint and surviving text: %s", len(payload.System), output)
	}
	if !strings.Contains(payload.System[0].Text, "x-anthropic-billing-header") {
		t.Fatalf("missing fingerprint: %s", output)
	}
	if payload.System[1].Text != "Keep this instruction" {
		t.Fatalf("unexpected surviving text: %q", payload.System[1].Text)
	}
	for _, block := range payload.System {
		if block.Type == "text" && strings.TrimSpace(block.Text) == "" {
			t.Fatalf("empty system text block: %s", output)
		}
	}
}
