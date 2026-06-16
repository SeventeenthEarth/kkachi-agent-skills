.PHONY: test test-prepare test-unit test-int test-e2e build install push-tag

GOCACHE ?= /tmp/kkachi-agent-skills-go-build
GOPATH ?= /tmp/kkachi-agent-skills-go-path
export GOCACHE
export GOPATH

BIN_DIR := bin
BINARY  := kkachi-agent-skills

test-prepare:
	@test -z "$$(gofmt -l main.go cmd internal tests)" || (gofmt -l main.go cmd internal tests && exit 1)
	go vet ./...
	go build -o /tmp/kkachi-agent-skills-test-build .
	go build -o /tmp/kkachi-hermes-skills-test-build ./cmd/kkachi-hermes-skills
	go test ./tests/docs_contract

test-unit:
	go test ./internal/skills/discovery ./internal/skills/install ./internal/skills/doctor ./internal/skills/graphsync ./internal/skills/kasstate ./internal/skills/projectinstall ./internal/skills/workflowcreator ./internal/skills/workflowregistry ./internal/skills/workflowrouting ./internal/skills/workflowtrigger

test-int:
	go test ./internal/skills/cli

test-e2e:
	go test ./tests/e2e

test: test-prepare test-unit test-int test-e2e

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) .

install:
	env -u GOPATH go install .

push-tag: build
	@set -e; \
	LOCAL_VER="v$$(./$(BIN_DIR)/$(BINARY) --version | awk '{print $$NF}')"; \
	echo ">> local binary version: $$LOCAL_VER"; \
	if ! git rev-parse -q --verify "refs/tags/$$LOCAL_VER" >/dev/null; then \
		echo "ERROR: local tag $$LOCAL_VER not found. Create it first:  git tag $$LOCAL_VER"; \
		exit 1; \
	fi; \
	ORIGIN_VER="$$(git ls-remote --tags origin 2>/dev/null | sed 's,.*refs/tags/,,' | grep -v '\^{}' | grep -E '^v[0-9]' | sort -V | tail -1)"; \
	if [ -z "$$ORIGIN_VER" ]; then echo "ERROR: could not read tags from origin (network/auth?)"; exit 1; fi; \
	echo ">> origin latest version: $$ORIGIN_VER"; \
	if [ "$$LOCAL_VER" = "$$ORIGIN_VER" ]; then \
		echo "FAIL-CLOSE: $$LOCAL_VER already on origin; bump CLIVersion in internal/skills/version/version.go"; \
		exit 1; \
	fi; \
	HIGHEST="$$(printf '%s\n%s\n' "$$LOCAL_VER" "$$ORIGIN_VER" | sort -V | tail -1)"; \
	if [ "$$HIGHEST" != "$$LOCAL_VER" ]; then \
		echo "FAIL-CLOSE: local $$LOCAL_VER is behind origin $$ORIGIN_VER"; \
		exit 1; \
	fi; \
	echo ">> OK: $$LOCAL_VER > $$ORIGIN_VER -- pushing tag"; \
	git push origin "$$LOCAL_VER"
