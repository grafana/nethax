.PHONY: deps deps-nethax deps-probe build build-nethax build-probe test


CUR_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
DIR_NETHAX = "$(CUR_DIR)/cmd/nethax"
DIR_PROBE = "$(CUR_DIR)/cmd/probe"

deps: deps-nethax deps-probe

deps-nethax:
	@cd ${DIR_NETHAX} && go mod download

deps-probe:
	@cd ${DIR_PROBE} && go mod download

build: build-nethax build-probe

build-nethax: deps-nethax
	@cd ${DIR_NETHAX} && go build -o "$(CUR_DIR)/bin"

build-probe: deps-probe
	@cd ${DIR_PROBE} && go build -o "$(CUR_DIR)/bin"

test:
	@go test ./...