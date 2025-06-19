CI := $(CI)

ifndef CI
# Check we've got the necessary tools installed...
EXECUTABLES := git go docker kind kubectl
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)), not found , $(error "No $(exec) in PATH")))
endif

# Set these to release new versions of the container
RUNNER_SEMVER := "0.0.2"
PROBE_SEMVER := "0.0.2"

ifdef CI
	RUNNER_VERSION := $(RUNNER_SEMVER)
	PROBE_VERSION := $(PROBE_SEMVER)
else
	WORKING_TREE_SHA := $(shell hack/get-worktree-sha.sh)
	RUNNER_VERSION := "$(RUNNER_SEMVER)-$(WORKING_TREE_SHA)"
	PROBE_VERSION := "$(PROBE_SEMVER)-$(WORKING_TREE_SHA)"
endif

RUNNER_IMAGE_PREFIX := "grafana/nethax-runner"
PROBE_IMAGE_PREFIX := "grafana/nethax-probe"

CUR_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
DIR_RUNNER := "$(CUR_DIR)/cmd/nethax"
DIR_PROBE := "$(CUR_DIR)/cmd/nethax-probe"

# Enable experimental testing/synctest package
export GOEXPERIMENT=synctest

build: build-runner build-probe

build-runner: deps-runner
	go build -ldflags="-X 'github.com/grafana/nethax/pkg/kubernetes.DefaultProbeImage=$(PROBE_IMAGE_PREFIX):$(PROBE_VERSION)'" -o "$(CUR_DIR)/bin" $(DIR_RUNNER)

build-probe: deps-probe
	go build -ldflags="-X 'github.com/grafana/nethax/pkg/kubernetes.DefaultProbeImage=$(PROBE_IMAGE_PREFIX):$(PROBE_VERSION)'" -o "$(CUR_DIR)/bin" $(DIR_PROBE)

.PHONY: clean
clean:
	@rm -f "$(CUR_DIR)/bin/"*

.PHONY: deps deps-runner deps-probe

# Default kind cluster name if not overridden
KIND_CLUSTER_NAME ?= "nethax"

deps:
	go mod download

.PHONY: docker-build docker-push docker-runner docker-probe
docker-build: docker-runner docker-probe

docker-push:
	@docker push $(RUNNER_IMAGE_PREFIX):$(RUNNER_VERSION)
	@docker push $(RUNNER_IMAGE_PREFIX):latest
	@docker push $(PROBE_IMAGE_PREFIX):$(PROBE_VERSION)
	@docker push $(PROBE_IMAGE_PREFIX):latest

docker-runner:
	@docker buildx build -f Dockerfile-runner \
		--build-arg PROBE_IMAGE=$(PROBE_IMAGE_PREFIX):$(PROBE_VERSION) \
		-t $(RUNNER_IMAGE_PREFIX):$(RUNNER_VERSION) -t $(RUNNER_IMAGE_PREFIX):latest \
		--platform="linux/amd64,linux/arm64,linux/ppc64le,linux/s390x" \
		.
ifndef CI
	@kind load docker-image --name $(KIND_CLUSTER_NAME) $(RUNNER_IMAGE_PREFIX):$(RUNNER_VERSION) || true
endif

docker-probe:
	@docker buildx build -f Dockerfile-probe \
		--build-arg PROBE_IMAGE=$(PROBE_IMAGE_PREFIX):$(PROBE_VERSION) \
		-t $(PROBE_IMAGE_PREFIX):$(PROBE_VERSION) -t $(PROBE_IMAGE_PREFIX):latest \
		--platform="linux/amd64,linux/arm64,linux/ppc64le,linux/s390x" \
		.
ifndef CI
	@kind load docker-image --name $(KIND_CLUSTER_NAME) $(PROBE_IMAGE_PREFIX):$(PROBE_VERSION) || true
endif

.PHONY: test
test:
	@go test -cover ./...

.PHONY: kind-init-oteldemo

# Re-initializes a kind cluster with the OTel demo
kind-init-oteldemo:
	@kind delete cluster --name $(KIND_CLUSTER_NAME) || true
	@kind create cluster --name $(KIND_CLUSTER_NAME)
	@kubectl --context "kind-$(KIND_CLUSTER_NAME)" create ns otel-demo
	@kubectl --context "kind-$(KIND_CLUSTER_NAME)" apply -n otel-demo -f https://raw.githubusercontent.com/open-telemetry/opentelemetry-demo/main/kubernetes/opentelemetry-demo.yaml || true
	@kubectl --context "kind-$(KIND_CLUSTER_NAME)" create cm -n otel-demo grafana-dashboards
	@kubectl --context "kind-$(KIND_CLUSTER_NAME)" replace -n otel-demo -f https://raw.githubusercontent.com/open-telemetry/opentelemetry-demo/main/kubernetes/opentelemetry-demo.yaml

.PHONY: run-example-test-plan
# Default test plan path if not overridden
TEST_PLAN ?= "$(CUR_DIR)example/OtelDemoTestPlan.yaml"
# Run the example test plan against KIND_CLUSTER_NAME
run-example-test-plan: docker-build
	@echo "Running test plan: $(TEST_PLAN)"
	@TMP_KUBECONFIG=$$(mktemp) && \
	kind get kubeconfig --name $(KIND_CLUSTER_NAME) > $$TMP_KUBECONFIG && \
	docker run --rm \
		--network host \
		--mount "type=bind,source=$$TMP_KUBECONFIG,target=/.kube/config,readonly" \
		--mount "type=bind,source=$(TEST_PLAN),target=/test-plan.yaml,readonly" \
		-e "KUBECONFIG=/.kube/config" \
		--user $(id -u):$(id -g) \
		$(RUNNER_IMAGE_PREFIX):$(RUNNER_VERSION) "execute-test" "-f" "/test-plan.yaml"; \
	rm -rf $$TMP_KUBECONFIG

.PHONY: checks
checks:
	go mod tidy --diff | grep ^ && exit 1 || true
	go vet -all ./...
	go tool staticcheck ./...
	go tool staticcheck -tests=false ./...
