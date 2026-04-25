SHELL := /bin/bash
PNPM := corepack pnpm

.PHONY: help install lint-spec lint-openapi lint-docs lint-fixtures lint-adapter-profiles generate-typescript-spec check-typescript-spec check-agentdesk-ts build-agentdesk-ts validate-examples test-go build-go build-cli check-typescript-sdk check-python-sdk check-sdk check-adapters check-examples check-ui build-ui smoke e2e simulation verify release-check dev-up dev-down run-server run-ui

help:
	@echo "Available targets:"
	@echo "  make install    - Install workspace dependencies with pnpm"
	@echo "  make verify     - Run spec lint, Go tests/build, UI checks/build, and smoke test"
	@echo "  make lint-openapi - Validate the OpenAPI control-plane contract"
	@echo "  make lint-docs - Validate local Markdown links"
	@echo "  make lint-fixtures - Validate fixture and example references"
	@echo "  make lint-adapter-profiles - Validate adapter conformance profiles"
	@echo "  make generate-typescript-spec - Regenerate TypeScript SDK spec types"
	@echo "  make check-typescript-spec - Check generated TypeScript spec types are current"
	@echo "  make check-agentdesk-ts - Build and test the TypeScript-first AgentDesk packages"
	@echo "  make validate-examples - Validate public examples through the Guild CLI"
	@echo "  make smoke      - Run the local API smoke test"
	@echo "  make e2e        - Run CLI, SDK, adapter, and replay e2e"
	@echo "  make simulation - Run the full one task/DRI/artifact/replay/commons simulation"
	@echo "  make release-check - Run full pre-release verification"
	@echo "  make run-server - Run the bootstrap control plane"
	@echo "  make run-ui     - Run the Next.js UI"
	@echo "  make dev-up     - Start optional local infrastructure"
	@echo "  make dev-down   - Stop optional local infrastructure"

install:
	$(PNPM) install --frozen-lockfile

lint-spec:
	$(PNPM) run lint:spec

lint-openapi:
	$(PNPM) run lint:openapi

lint-docs:
	$(PNPM) run lint:docs

lint-fixtures:
	$(PNPM) run lint:fixtures

lint-adapter-profiles:
	$(PNPM) run lint:adapter-profiles

generate-typescript-spec:
	$(PNPM) run generate:typescript-spec

check-typescript-spec:
	$(PNPM) run check:typescript-spec

build-agentdesk-ts:
	$(PNPM) run build:agentdesk-ts

check-agentdesk-ts:
	$(PNPM) run check:agentdesk-ts

validate-examples:
	go run ./cli/cmd/guild validate --kind taskpack --file spec/examples/taskpack.example.json
	go run ./cli/cmd/guild validate --kind dri-binding --file spec/examples/dri-binding.example.json
	go run ./cli/cmd/guild validate --kind artifact --file spec/examples/artifact.example.json
	go run ./cli/cmd/guild validate --kind promotion-record --file spec/examples/promotion-record.example.json
	go run ./cli/cmd/guild validate --kind governance-policy --file spec/examples/governance-policy.example.json
	go run ./cli/cmd/guild validate --kind approval-request --file spec/examples/approval-request.example.json
	go run ./cli/cmd/guild validate --kind promotion-gate --file spec/examples/promotion-gate.example.json
	go run ./cli/cmd/guild validate --kind commons-entry --file spec/examples/commons-entry.example.json
	go run ./cli/cmd/guild validate --kind replay-bundle --file spec/examples/replay-bundle.example.json
	go run ./cli/cmd/guild validate --kind workspace-constitution --file spec/examples/workspace-constitution.example.json
	go run ./cli/cmd/guild validate --kind context-pack --file spec/examples/context-pack.example.json
	go run ./cli/cmd/guild validate --kind preflight-decision --file spec/examples/preflight-decision.example.json

test-go:
	go test ./...

build-go:
	go build ./server/cmd/guildd
	rm -f guildd

build-cli:
	go build ./cli/cmd/guild
	rm -f guild

check-typescript-sdk:
	$(PNPM) --dir sdk/typescript run check

check-python-sdk:
	python3 -m compileall -q sdk/python/src
	PYTHONPATH=sdk/python/src python3 -m unittest discover -s sdk/python/tests

check-sdk: check-typescript-sdk check-python-sdk

check-adapters:
	$(PNPM) --dir adapters/typescript run check
	$(PNPM) --dir adapters/mcp run check
	$(PNPM) --dir adapters/a2a run check
	$(PNPM) --dir adapters/langgraph run check

check-examples:
	$(PNPM) --dir examples/typescript-adapter-core run check

check-ui:
	$(PNPM) --dir ui run check

build-ui:
	$(PNPM) --dir ui run build

smoke:
	./scripts/smoke.sh

e2e:
	./scripts/e2e.sh

simulation:
	./scripts/simulation.sh

release-check:
	./scripts/release-check.sh

verify: install lint-spec lint-openapi lint-docs lint-fixtures lint-adapter-profiles check-typescript-spec check-agentdesk-ts validate-examples test-go build-go build-cli check-sdk check-adapters check-examples check-ui build-ui smoke

dev-up:
	docker compose -f deploy/docker-compose.local.yml up -d

dev-down:
	docker compose -f deploy/docker-compose.local.yml down -v

run-server:
	go run ./server/cmd/guildd

run-ui:
	$(PNPM) --dir ui run dev
