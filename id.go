package main

import (
	"crypto/sha256"
	"encoding/base64"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const Entropy int = 43
const Alphabet string = "abcdefghijklmnopqrstuvwxyz0123456789"

func GenerateVerifier() (string, error) {
	return gonanoid.Generate(Alphabet, Entropy)
}
func HashCodeVerifier(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
