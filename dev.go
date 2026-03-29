package main

import (
	"fmt"
	"log"
	"os"
)

const Samples = "samples/"

func tinker() {
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
