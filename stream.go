package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

func relay_stream(w io.Writer, upstream io.Reader, flush func()) error {
	reader := bufio.NewReader(upstream)
	var event string
	var data strings.Builder
	var size int

	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			if _, writeErr := io.WriteString(w, line); writeErr != nil {
				return writeErr
			}
			flush()
			field := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			switch {
			case strings.HasPrefix(field, "event:"):
				event = strings.TrimSpace(strings.TrimPrefix(field, "event:"))
			case strings.HasPrefix(field, "data:"):
				data.WriteString(strings.TrimSpace(strings.TrimPrefix(field, "data:")))
				size += len(field)
			case field == "":
				inspect_stream_event(event, data.String(), size)
				event = ""
				data.Reset()
				size = 0
			}
		}
		if err == io.EOF {
			if data.Len() != 0 {
				inspect_stream_event(event, data.String(), size)
			}
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func inspect_stream_event(event, data string, size int) {
	if data == "" {
		return
	}
	var payload struct {
		Type         string          `json:"type"`
		Index        json.RawMessage `json:"index"`
		Message      json.RawMessage `json:"message"`
		ContentBlock json.RawMessage `json:"content_block"`
		Delta        json.RawMessage `json:"delta"`
		Usage        json.RawMessage `json:"usage"`
	}
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		log.Printf("invalid stream JSON: event=%q bytes=%d error=%v", event, size, err)
		return
	}
	var missing string
	switch payload.Type {
	case "content_block_start":
		if !json.Valid(payload.ContentBlock) || len(payload.ContentBlock) == 0 || payload.ContentBlock[0] != '{' {
			missing = "content_block"
		}
	case "content_block_delta", "message_delta":
		if !json.Valid(payload.Delta) || len(payload.Delta) == 0 || payload.Delta[0] != '{' {
			missing = "delta"
		}
	}
	if missing != "" || event != "" && event != payload.Type {
		log.Printf("unexpected stream event: event=%q type=%q bytes=%d missing=%s index=%s", event, payload.Type, size, missing, safe_index(payload.Index))
	}
}

func safe_index(index json.RawMessage) string {
	if len(index) == 0 {
		return "absent"
	}
	var n int
	if err := json.Unmarshal(index, &n); err != nil {
		return fmt.Sprintf("invalid(%d bytes)", len(index))
	}
	return fmt.Sprint(n)
}
