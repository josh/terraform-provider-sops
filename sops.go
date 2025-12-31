package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type SopsEncryptOptions struct {
	AgeRecipients []string
	OutputType    string
	OutputIndent  *int64
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

	args := []string{}

	if opts.OutputIndent != nil {
		args = append(args, "--indent", fmt.Sprintf("%d", *opts.OutputIndent))
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
		return nil, fmt.Errorf("sops encrypt failed: %s", stderr.String())
	}

	return stdout.Bytes(), nil
}

type SopsDecryptOptions struct {
	AgeIdentityPath  string
	AgeIdentityValue string
}

func decryptWithSops(ctx context.Context, encryptedJSON []byte, opts SopsDecryptOptions) ([]byte, error) {
	args := []string{"decrypt", "--input-type", "json", "--output-type", "json", "/dev/stdin"}
	cmd := exec.CommandContext(ctx, sopsBinary, args...)
	cmd.Stdin = bytes.NewReader(encryptedJSON)

	cmd.Env = os.Environ()
	if opts.AgeIdentityValue != "" {
		cmd.Env = append(cmd.Env, "SOPS_AGE_KEY="+opts.AgeIdentityValue)
	} else if opts.AgeIdentityPath != "" {
		cmd.Env = append(cmd.Env, "SOPS_AGE_KEY_FILE="+opts.AgeIdentityPath)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("sops decrypt failed: %s", stderr.String())
	}

	return stdout.Bytes(), nil
}
