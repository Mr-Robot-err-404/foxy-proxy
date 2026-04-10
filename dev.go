package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

const Samples = "samples/"

func tinker() {
	test_strip_third_party()
}

func sanitize_sample() {
	path := make_path("partial.json")
	b, err := os.ReadFile(path)

	if err != nil {
		log.Fatal(err)
	}
	output, err := sanitize_payload(b)
	if err != nil {
		log.Fatal(err)
	}
	path = make_path("output.json")
	os.WriteFile(path, output, 0644)
}

func make_path(name string) string {
	return fmt.Sprintf("%s%s", Samples, name)
}

func test_strip_third_party() {
	path := make_path("third_party.json")
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var item SystemItem
	if err := json.Unmarshal(b, &item); err != nil {
		log.Fatal(err)
	}

	stripped := strip_system_prompt(item.Text)
	if contains_keyword(stripped) {
		log.Fatal("keyword stripping failed")
	}

	fmt.Printf("stripping works: removed %d chars\n", len(item.Text)-len(stripped))
	if stripped != item.Text {
		fmt.Println(strings.Repeat("-", 40))
		fmt.Println(stripped)
	}
}
