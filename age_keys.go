package main

import (
	"fmt"
	"strings"

	"filippo.io/age"
)

func generateAgeKeyPair() (privateKey string, publicKey string, err error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate age identity: %w", err)
	}

	return identity.String(), identity.Recipient().String(), nil
}

func deriveAgePublicKey(privateKey string) (string, error) {
	identity, err := age.ParseX25519Identity(strings.TrimSpace(privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse age private key: %w", err)
	}

	return identity.Recipient().String(), nil
}
