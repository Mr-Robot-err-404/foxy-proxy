package main

import (
	"encoding/json"
	"strings"
)

type SystemItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type ToolsItem struct {
	Name string `json:"name"`
}

type Payload map[string]json.RawMessage
type Transformation = func(m Payload) error

var (
	EmptyArray  = json.RawMessage(`[]`)
	EmptyObject = json.RawMessage(`{}`)
	EmptyString = json.RawMessage(`""`)
)

var System = []string{"opencode", "<directories>", "</directories>", "Here is some useful information about the environment you are running in:"}

var ToolsRemapping = map[string]string{
	"todowrite": "taskwrite",
}

var Transformations = []Transformation{
	func(m Payload) error {
		return transform_json(m, "system", system_transform)
	},
	func(m Payload) error {
		return transform_json(m, "tools", tools_transform)
	},
	func(m Payload) error {
		return ensure_exists(m, "metadata", EmptyObject)
	},
	func(m Payload) error {
		return ensure_exists(m, "tools", EmptyArray)
	},
}
var Fingerprint = SystemItem{
	Type: "text",
	Text: "x-anthropic-billing-header: cc_version=2.1.81.df2; cc_entrypoint=cli; cch=c90fe;",
}

var RequiredHeaders = map[string]string{
	"User-Agent":                                "claude-cli/2.1.81 (external, cli)",
	"X-Stainless-Arch":                          "arm64",
	"X-Stainless-Lang":                          "js",
	"X-Stainless-OS":                            "MacOS",
	"X-Stainless-Package-Version":               "0.74.0",
	"X-Stainless-Retry-Count":                   "0",
	"X-Stainless-Runtime":                       "node",
	"X-Stainless-Runtime-Version":               "v24.3.0",
	"X-Stainless-Timeout":                       "600",
	"anthropic-beta":                            "claude-code-20250219,oauth-2025-04-20,interleaved-thinking-2025-05-14,redact-thinking-2026-02-12,context-management-2025-06-27,prompt-caching-scope-2026-01-05,advanced-tool-use-2025-11-20,effort-2025-11-24",
	"anthropic-dangerous-direct-browser-access": "true",
	"anthropic-version":                         "2023-06-01",
	"x-app":                                     "cli",
}
var ExcludedHeaders = []string{"x-api-key"}

func sanitize_payload(b []byte) ([]byte, error) {
	var payload map[string]json.RawMessage
	err := json.Unmarshal(b, &payload)

	if err != nil {
		return []byte{}, err
	}
	for _, transform := range Transformations {
		err := transform(payload)
		if err != nil {
			return []byte{}, err
		}
	}
	return json.Marshal(payload)
}

func transform_json[T any](m Payload, key string, transform func(T) T) error {
	var value T
	if data, exists := m[key]; exists {
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
	}
	return set_json(m, key, transform(value))
}

func ensure_exists(m Payload, key string, fallback json.RawMessage) error {
	_, ok := m[key]
	if ok {
		return nil
	}
	return set_json(m, key, fallback)
}

func set_json[T any](m Payload, key string, value T) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m[key] = json.RawMessage(bytes)
	return nil
}

func tools_transform(tools []map[string]json.RawMessage) []map[string]json.RawMessage {
	for _, tool := range tools {
		var name string
		json.Unmarshal(tool["name"], &name)

		if renamed, ok := ToolsRemapping[name]; ok {
			tool["name"] = json.RawMessage(`"` + renamed + `"`)
		}
	}
	return tools
}

func system_transform(system []SystemItem) []SystemItem {
	updated := []SystemItem{}

	for _, current := range system {
		current.Text = strip_system_prompt(current.Text)
		updated = append(updated, current)
	}
	if missing_system_item(updated) {
		correction := []SystemItem{Fingerprint}
		updated = append(correction, updated...)
	}
	return updated
}

func missing_system_item(items []SystemItem) bool {
	for _, item := range items {
		if item.Text == Fingerprint.Text {
			return false
		}
	}
	return true
}

func strip_system_prompt(prompt string) string {
	s := strings.Builder{}

	for line := range strings.SplitSeq(prompt, "\n") {
		if contains_keyword(line, System) {
			continue
		}
		s.WriteString(line)
		s.WriteString("\n")
	}
	return strings.TrimSuffix(s.String(), "\n")
}

func contains_keyword(line string, keywords []string) bool {
	for _, k := range keywords {
		if strings.Contains(line, k) {
			return true
		}
	}
	return false
}
