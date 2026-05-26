.PHONY: test test-prepare test-unit test-int test-e2e

GOCACHE ?= /tmp/kkachi-hermes-skills-go-build
GOPATH ?= /tmp/kkachi-hermes-skills-go-path
export GOCACHE
export GOPATH

test-prepare:
	go test ./...
	go vet ./...
	go build -o /tmp/kkachi-hermes-skills-test-build ./cmd/kkachi-hermes-skills

test-unit:
	go test ./internal/skills/discovery ./internal/skills/install

test-int:
	go test ./internal/skills/cli

test-e2e:
	go test ./tests/e2e

test: test-prepare test-unit test-int test-e2e
