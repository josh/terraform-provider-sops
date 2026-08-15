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

// parseAgeIdentity accepts a bare AGE-SECRET-KEY-1... string or age-keygen
// file content, the same shapes SOPS accepts for identities.
func parseAgeIdentity(privateKey string) (*age.X25519Identity, error) {
	var identity *age.X25519Identity

	for _, line := range strings.Split(privateKey, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Plugin identity recipients cannot be derived offline.
		if strings.HasPrefix(line, "AGE-PLUGIN-") {
			return nil, fmt.Errorf("age plugin identities are not supported: the public key of a plugin identity cannot be derived from the identity string")
		}

		parsed, err := age.ParseX25519Identity(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse age private key: %w", err)
		}

		if identity != nil {
			return nil, fmt.Errorf("multiple age identities found; expected exactly one")
		}
		identity = parsed
	}

	if identity == nil {
		return nil, fmt.Errorf("no age private key found in input")
	}

	return identity, nil
}

func deriveAgePublicKey(privateKey string) (string, error) {
	identity, err := parseAgeIdentity(privateKey)
	if err != nil {
		return "", err
	}

	return identity.Recipient().String(), nil
}
