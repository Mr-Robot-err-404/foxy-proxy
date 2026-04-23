package main

import (
	"encoding/json"
	"log"
	"os"
)

func tinker() {
	inject_tools("item.json", "arms_race/payload.json")
	sanitize_sample("arms_race/payload.json", "arms_race/sanitized.json")
}

func sanitize_sample(path string, save string) {
	b, err := os.ReadFile(path)

	if err != nil {
		log.Fatal(err)
	}
	output, err := sanitize_payload(b)
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile(save, output, 0644)
}

func inject_tools(tools_path string, save string) {
	template, err := os.ReadFile("required/required.json")
	if err != nil {
		log.Fatal(err)
	}
	tools, err := os.ReadFile(tools_path)
	if err != nil {
		log.Fatal(err)
	}
	var payload map[string]json.RawMessage

	if err := json.Unmarshal(template, &payload); err != nil {
		log.Fatal(err)
	}
	payload["tools"] = json.RawMessage(tools)

	out, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile(save, out, 0644)
}
