package main

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestRelayStreamPreservesLargeEventsAndFollowingFrames(t *testing.T) {
	large := `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"` + strings.Repeat("a", 1024*1024) + `"}}`
	input := "event: content_block_delta\ndata: " + large + "\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"
	var output bytes.Buffer
	flushes := 0
	if err := relay_stream(&output, strings.NewReader(input), func() { flushes++ }); err != nil {
		t.Fatal(err)
	}
	if output.String() != input {
		t.Fatalf("relay did not preserve %d-byte stream", len(input))
	}
	if flushes < 4 {
		t.Fatalf("expected incremental flushing, got %d flushes", flushes)
	}
}

func TestInspectStreamEventLogsShapeWithoutContent(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)

	inspect_stream_event("content_block_delta", `{"type":"content_block_delta","delta":null,"text":"private prompt"}`, 80)
	if !strings.Contains(output.String(), "missing=delta") {
		t.Fatalf("missing shape diagnostic: %s", output.String())
	}
	if strings.Contains(output.String(), "private prompt") {
		t.Fatal("logged response content")
	}
	output.Reset()
	inspect_stream_event("content_block_delta", `{"type":"content_block_delta","delta":{"type":"text_delta","text":"private prompt"}}`, 90)
	if output.Len() != 0 {
		t.Fatalf("unexpected diagnostic: %s", output.String())
	}
}
