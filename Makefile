# SPDX-FileCopyrightText: 2026 Bonial International GmbH
# SPDX-License-Identifier: Apache-2.0

.PHONY: help lint reuse-lint md-lint go-lint test build install-tools clean

help:
	@echo "Available targets:"
	@echo "  make lint           - Run reuse-lint + md-lint + go-lint"
	@echo "  make reuse-lint     - Run REUSE lint only"
	@echo "  make md-lint        - Run markdown lint only"
	@echo "  make go-lint        - Run golangci-lint"
	@echo "  make test           - Run go test ./... -race -cover"
	@echo "  make build          - Build the demo binary into ./demo"
	@echo "  make install-tools  - Install lint tooling into a local venv"
	@echo "  make clean          - Remove build artifacts"

lint: reuse-lint md-lint go-lint

reuse-lint:
	@command -v reuse >/dev/null 2>&1 || { \
	  echo "reuse not found; run 'make install-tools' first" >&2; exit 1; \
	}
	reuse lint

md-lint:
	@command -v npx >/dev/null 2>&1 || { \
	  echo "npx not found; install Node.js first" >&2; exit 1; \
	}
	npx --yes markdownlint-cli2 \
	  "**/*.md" \
	  "!LICENSES/**" \
	  "!.venv-reuse/**" \
	  "!.worktrees/**" \
	  "!.claude/**" \
	  "!.superpowers/**" \
	  "!node_modules/**" \
	  "!vendor/**"

go-lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
	  echo "golangci-lint not found; install from https://golangci-lint.run/" >&2; exit 1; \
	}
	golangci-lint run ./...

test:
	go test ./... -race -cover

build:
	go build -trimpath -o demo ./

install-tools:
	python3 -m venv .venv-reuse
	.venv-reuse/bin/pip install --quiet 'reuse>=4' charset-normalizer
	@echo ""
	@echo "Add the venv to PATH for this shell:"
	@echo "  export PATH=\"$$(pwd)/.venv-reuse/bin:\$$PATH\""
	@echo ""
	@echo "Also install (once):"
	@echo "  - Go 1.24+       (https://go.dev/dl/)"
	@echo "  - golangci-lint  (https://golangci-lint.run/)"
	@echo "  - Node.js        (for markdownlint-cli2 via npx)"

clean:
	rm -f ./demo
	rm -rf ./dist
