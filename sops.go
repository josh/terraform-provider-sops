package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func formatSopsStderr(stderr string) string {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return ""
	}
	return ": " + stderr
}

type SopsEncryptOptions struct {
	AgeRecipients     []string
	OutputType        string
	OutputIndent      *int64
	UnencryptedSuffix *string
	EncryptedSuffix   *string
	UnencryptedRegex  *string
	EncryptedRegex    *string
}

func encryptWithSops(ctx context.Context, input map[string]interface{}, opts SopsEncryptOptions) ([]byte, error) {
	if len(opts.AgeRecipients) == 0 {
		return nil, fmt.Errorf("at least one age recipient must be provided")
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input to JSON: %w", err)
	}

	outputType := opts.OutputType
	if outputType == "" {
		outputType = "json"
	}

	args := []string{"--config", "/dev/null"}

	if opts.OutputIndent != nil {
		args = append(args, "--indent", fmt.Sprintf("%d", *opts.OutputIndent))
	}

	if opts.UnencryptedSuffix != nil {
		args = append(args, "--unencrypted-suffix", *opts.UnencryptedSuffix)
	}

	if opts.EncryptedSuffix != nil {
		args = append(args, "--encrypted-suffix", *opts.EncryptedSuffix)
	}

	if opts.UnencryptedRegex != nil {
		args = append(args, "--unencrypted-regex", *opts.UnencryptedRegex)
	}

	if opts.EncryptedRegex != nil {
		args = append(args, "--encrypted-regex", *opts.EncryptedRegex)
	}

	args = append(args, "--encrypt", "--input-type", "json", "--output-type", outputType, "/dev/stdin")
	cmd := exec.CommandContext(ctx, sopsBinary, args...)
	cmd.Stdin = bytes.NewReader(inputJSON)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SOPS_AGE_RECIPIENTS="+strings.Join(opts.AgeRecipients, ","))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("sops encrypt failed: %w%s", err, formatSopsStderr(stderr.String()))
	}

	return stdout.Bytes(), nil
}

func expandTilde(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}

	return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
}

type SopsDecryptOptions struct {
	AgeIdentityPath  string
	AgeIdentityValue string
	InputType        string
}

func decryptWithSops(ctx context.Context, encryptedData []byte, opts SopsDecryptOptions) ([]byte, error) {
	inputType := opts.InputType
	if inputType == "" {
		return nil, fmt.Errorf("input type is required")
	}

	args := []string{"--config", "/dev/null", "decrypt", "--input-type", inputType, "--output-type", "json", "/dev/stdin"}
	cmd := exec.CommandContext(ctx, sopsBinary, args...)
	cmd.Stdin = bytes.NewReader(encryptedData)

	cmd.Env = os.Environ()
	if opts.AgeIdentityValue != "" {
		cmd.Env = append(cmd.Env, "SOPS_AGE_KEY="+opts.AgeIdentityValue)
	} else if opts.AgeIdentityPath != "" {
		identityPath, err := expandTilde(opts.AgeIdentityPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve age identity file path %q: %w", opts.AgeIdentityPath, err)
		}
		if _, err := os.Stat(identityPath); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("age identity file not found: %s", identityPath)
			}
			return nil, fmt.Errorf("failed to access age identity file %s: %w", identityPath, err)
		}
		cmd.Env = append(cmd.Env, "SOPS_AGE_KEY_FILE="+identityPath)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("sops decrypt failed: %w%s", err, formatSopsStderr(stderr.String()))
	}

	return stdout.Bytes(), nil
}
