.PHONY: test test-prepare test-unit test-int test-e2e

GOCACHE ?= /tmp/kkachi-hermes-skills-go-build
GOPATH ?= /tmp/kkachi-hermes-skills-go-path
export GOCACHE
export GOPATH

test-prepare:
	@test -z "$$(gofmt -l cmd internal tests)" || (gofmt -l cmd internal tests && exit 1)
	go vet ./...
	go build -o /tmp/kkachi-hermes-skills-test-build ./cmd/kkachi-hermes-skills
	go test ./tests/docs_contract

test-unit:
	go test ./internal/skills/discovery ./internal/skills/install ./internal/skills/doctor ./internal/skills/kasstate

test-int:
	go test ./internal/skills/cli

test-e2e:
	go test ./tests/e2e

test: test-prepare test-unit test-int test-e2e
