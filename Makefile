.PHONY: deps deps-runner deps-probe build build-runner build-probe test docker docker-runner docker-probe

# Set these to release new versions of the container
RUNNER_SEMVER := "0.1.0"
PROBE_SEMVER := "0.1.0"

CI := $(CI)

ifdef CI
	RUNNER_VERSION := ${RUNNER_SEMVER}
	PROBE_VERSION := ${PROBE_SEMVER}
else
	RUNNER_VERSION := "${RUNNER_SEMVER}-$(shell date +%s)"
	PROBE_VERSION := "${PROBE_SEMVER}-$(shell date +%s)"
endif

CUR_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
DIR_RUNNER := "$(CUR_DIR)/cmd/runner"
DIR_PROBE := "$(CUR_DIR)/cmd/probe"

build: build-runner build-probe

build-runner: deps-runner
	@cd ${DIR_RUNNER} && go build -o "$(CUR_DIR)/bin"

build-probe: deps-probe
	@cd ${DIR_PROBE} && go build -o "$(CUR_DIR)/bin"

clean:
	@rm -f "$(CUR_DIR)/bin/"*

deps: deps-runner deps-probe

deps-runner:
	@cd ${DIR_RUNNER} && go mod download

deps-probe:
	@cd ${DIR_PROBE} && go mod download


docker: docker-runner docker-probe

docker-runner:
	@docker build -f Dockerfile-runner -t nethax-runner:${RUNNER_VERSION} .
ifndef CI
	@kind load docker-image nethax-runner:${RUNNER_VERSION} || true
endif

docker-probe:
	@docker build -f Dockerfile-probe -t nethax-probe:${PROBE_VERSION} .
ifndef CI
	@kind load docker-image nethax-probe:${RUNNER_VERSION} || true
endif

test:
	@go test ./...
