

ENVTEST_K8S_VERSION = 1.21

# Linting
# Dependency versions
GOLANGCI_VERSION = 1.51.1
MOCKERY_VERSION = v2.28.1
GOLANGCI_BIN_NAME = golangci-lint
GOLANGCI_BIN_EXEC = ${GOLANGCI_BIN_NAME}

bin/golangci-lint-install:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./${GOLANGCI_BIN_EXEC} v${GOLANGCI_VERSION}

.PHONY: lint lint-fix
lint:
	${GOLANGCI_BIN_EXEC} run --timeout=5m

.PHONY: lint-fix
lint-fix:
	${GOLANGCI_BIN_EXEC} run --fix --timeout=5m

.PHONY: api
api: controller-gen ## Generate API objects
	${CONTROLLER_GEN} object paths="./api/v1alpha1/..."

.PHONY: crds
crds: controller-gen ## Generate CRDs
	${CONTROLLER_GEN} crd paths="./api/v1alpha1/..." output:crd:artifacts:config=charts/identity-manager/crds

.PHONY: gen
gen:
	go generate ./...

.PHONY: docker-mkdocs-serve
docker-mkdocs-serve:
	docker run --rm -it -p 8000:8000 -v $(shell pwd)/docs:/docs squidfunk/mkdocs-material serve

.PHONY: gen-crd-docs
gen-crd-docs:
	go install github.com/ahmetb/gen-crd-api-reference-docs@latest
	gen-crd-api-reference-docs \
			-template-dir ./hack/api-docs \
			-config ./hack/api-docs/config.json \
			-api-dir github.com/invisibl-cloud/identity-manager/api/v1alpha1 \
			-out-file docs/api-specification.md

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

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11.1)

ENVTEST = $(shell pwd)/bin/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: crds fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" CRDS_DIR=charts/identity-manager/crds go test ./... -v -coverprofile cover.out
