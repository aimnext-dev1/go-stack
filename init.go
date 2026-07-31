package main

import (
	"fmt"
	"os"
	"strings"
)

const stackEnvTemplate = `# docker compose project name (-p) / backup filename prefix
STACK_NAME=

# per-environment compose spec filenames (relative to the stack.env folder)
COMPOSE_FILE_LOCAL=docker-compose.yml
COMPOSE_FILE_DEV=docker-compose.yml
COMPOSE_FILE_PROD=docker-compose.yml

# (optional) shared base compose filename. When set, runs -f <base> -f <env file> combined
# COMPOSE_BASE_FILE=docker-compose.yml

# (optional) per-environment env file paths
# ENV_FILE_LOCAL=
# ENV_FILE_DEV=
# ENV_FILE_PROD=

# (optional) force container runtime: docker | podman (auto-detected if unset)
# GO_STACK_CONTAINER=docker
`

func cmdInit(args []string) error {
	if _, err := os.Stat("stack.env"); err == nil {
		return fmt.Errorf("stack.env already exists in this folder. Remove it first to recreate.")
	}
	if err := os.WriteFile("stack.env", []byte(stackEnvTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create stack.env: %w", err)
	}
	if err := ensureGitignored("stack.env"); err != nil {
		return err
	}
	redLog("stack.env created. Fill in the values, then run 'go-stack up'.")
	return nil
}

func ensureGitignored(entry string) error {
	data, err := os.ReadFile(".gitignore")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}
	content := string(data)
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += entry + "\n"
	return os.WriteFile(".gitignore", []byte(content), 0644)
}
