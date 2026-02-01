# Capture image tag from git branch name
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null || true)
ifeq (,$(GIT_BRANCH))
TAG = latest
else ifeq (master, $(GIT_BRANCH))
TAG = latest
else ifeq (HEAD, $(GIT_BRANCH))
TAG = $(shell git describe --abbrev=0 --tags $(shell git rev-list --abbrev-commit --tags --max-count=1) 2> /dev/null || true)
else
TAG = $(GIT_BRANCH)
endif

# Image URL to use all building/pushing image targets
IMG ?= netrisai/netris-operator:$(TAG)
# Produce CRDs
CRD_OPTIONS ?= "crd"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
ENVTEST = $(shell pwd)/bin/setup-envtest
ENVTEST_K8S_VERSION = 1.29
test: generate fmt vet manifests setup-envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(shell pwd)/bin -p path)" go test ./... -coverprofile cover.out

setup-envtest: ## Download setup-envtest locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./controllers/..." paths="./api/..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.20.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.7)

# go-get-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

release: generate fmt vet manifests kustomize
	$(KUSTOMIZE) build config/crd > deploy/netris-operator.crds.yaml
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > deploy/netris-operator.yaml

pip-install-reqs:
	pip3 install yq pyyaml

helm: generate fmt vet manifests pip-install-reqs
	mkdir -p deploy/charts/netris-operator/crds/
	cp config/crd/bases/* deploy/charts/netris-operator/crds/
	echo "{{- if .Values.rbac.create -}}" > deploy/charts/netris-operator/templates/rbac.yaml
	for i in $(shell yq -y .resources config/rbac/kustomization.yaml | awk {'print $$2'});\
	do echo "---" >> deploy/charts/netris-operator/templates/rbac.yaml && \
	scripts/rbac-helm-template.py config/rbac/$${i} | yq -y . >> deploy/charts/netris-operator/templates/rbac.yaml;\
	done
	echo "{{- end }}" >> deploy/charts/netris-operator/templates/rbac.yaml

helm-push: helm
	@{ \
	set -e ;\
	HELM_CHART_GEN_TMP_DIR=$$(mktemp -d) ;\
	git clone git@github.com:netrisai/charts.git --depth 1 $$HELM_CHART_GEN_TMP_DIR ;\
	if [[ -z "$${HELM_CHART_REPO_COMMIT_MSG}" ]]; then HELM_CHART_REPO_COMMIT_MSG=Update-$$(date -u +'%Y-%m-%d_%H:%M:%S'); fi ;\
	rm -rf $$HELM_CHART_GEN_TMP_DIR/charts/netris-operator ;\
	cp -r deploy/charts $$HELM_CHART_GEN_TMP_DIR ;\
	cd $$HELM_CHART_GEN_TMP_DIR ;\
	git add charts && git commit -m $$HELM_CHART_REPO_COMMIT_MSG && git push -u origin main ;\
	rm -rf $$HELM_CHART_GEN_TMP_DIR ;\
	}

# Unit tests
unit-test:
	go test -short $$(go list ./... | grep -v /e2e) --cover

# E2E Testing with kind cluster
E2E_CLUSTER_NAME ?= netris-e2e
E2E_IMG ?= netrisai/netris-operator:e2e

.PHONY: e2e-setup e2e-teardown e2e-test e2e

e2e-setup: ## Create kind cluster and deploy netris-controller
	@echo "Creating kind cluster..."
	@if ! kind get clusters | grep -q "^$(E2E_CLUSTER_NAME)$$"; then \
		kind create cluster --name $(E2E_CLUSTER_NAME) --wait 60s; \
	else \
		echo "Cluster $(E2E_CLUSTER_NAME) already exists, skipping creation"; \
	fi
	kubectl cluster-info --context kind-$(E2E_CLUSTER_NAME)
	@echo "Adding netrisai Helm repo..."
	helm repo add netrisai https://netrisai.github.io/charts || true
	helm repo update
	@echo "Installing netris-controller..."
	helm upgrade --install netris-controller netrisai/netris-controller -f hack/e2e-values.yaml --wait --timeout 10m
	@echo "Building and loading netris-operator image..."

e2e-build: kustomize
	docker build . -t $(E2E_IMG) --build-arg SKIP_TEST=true
	kind load docker-image $(E2E_IMG) --name $(E2E_CLUSTER_NAME)

e2e-deploy: e2e-build
	@echo "Installing CRDs..."
	$(KUSTOMIZE) build config/crd | kubectl apply -f -
	@echo "Deploying netris-operator..."
	$(KUSTOMIZE) build config/e2e | kubectl apply -f -
	@echo "Waiting for operator to be ready..."
	kubectl wait --for=condition=available --timeout=120s deployment/netris-operator-controller-manager -n netris-operator-system
	@echo "E2E environment ready"

e2e-teardown: ## Delete kind cluster
	kind delete cluster --name $(E2E_CLUSTER_NAME)

e2e-test: ## Run e2e tests against the kind cluster
	@echo "Running e2e tests..."
	go test ./e2e/... --ginkgo.slow-spec-threshold=60s --ginkgo.v -v -count=1

e2e: e2e-setup e2e-deploy e2e-test e2e-teardown ## Run full e2e test cycle
