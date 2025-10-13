# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

ENSURE_GARDENER_MOD         := $(shell go get github.com/gardener/gardener@$$(go list -m -f "{{.Version}}" github.com/gardener/gardener))
GARDENER_HACK_DIR           := $(shell go list -m -f "{{.Dir}}" github.com/gardener/gardener)/hack
EXTENSION_PREFIX            := gardener-extension
NAME                        := image-rewriter
ADMISSION_NAME              := $(NAME)-admission
IMAGE                       := europe-docker.pkg.dev/gardener-project/public/gardener/extensions/image-rewriter
REPO_ROOT                   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR                    := $(REPO_ROOT)/hack
VERSION                     := $(shell cat "$(REPO_ROOT)/VERSION")
EFFECTIVE_VERSION           := $(VERSION)-$(shell git rev-parse HEAD)
IMAGE_TAG                   := $(EFFECTIVE_VERSION)
LD_FLAGS                    := "-w $(shell bash $(GARDENER_HACK_DIR)/get-build-ld-flags.sh k8s.io/component-base $(REPO_ROOT)/VERSION $(NAME))"
PARALLEL_E2E_TESTS          := 3

ifneq ($(strip $(shell git status --porcelain 2>/dev/null)),)
	EFFECTIVE_VERSION := $(EFFECTIVE_VERSION)-dirty
endif

#########################################
# Tools                                 #
#########################################

TOOLS_DIR := $(HACK_DIR)/tools
include $(GARDENER_HACK_DIR)/tools.mk

#################################################################
# Rules related to binary build, Docker image build and release #
#################################################################

.PHONY: install
install:
	@LD_FLAGS=$(LD_FLAGS) \
	bash $(GARDENER_HACK_DIR)/install.sh ./cmd/...

.PHONY: docker-login
docker-login:
	@gcloud auth activate-service-account --key-file .kube-secrets/gcr/gcr-readwrite.json

.PHONY: docker-image
docker-image:
	@docker build --build-arg EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) -t $(IMAGE):$(IMAGE_TAG) -f Dockerfile -m 6g --target $(NAME) .

#####################################################################
# Rules for verification, formatting, linting, testing and cleaning #
#####################################################################

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: clean
clean:
	@$(shell find ./example -type f -name "controller-registration.yaml" -exec rm '{}' \;)
	@bash $(GARDENER_HACK_DIR)/clean.sh ./cmd/... ./pkg/...

.PHONY: check-generate
check-generate:
	@bash $(GARDENER_HACK_DIR)/check-generate.sh $(REPO_ROOT)

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT) $(HELM) $(YQ)
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/... ./test/...
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check-charts.sh ./charts

.PHONY: generate
generate: $(CONTROLLER_GEN) $(EXTENSION_GEN) $(GEN_CRD_API_REFERENCE_DOCS) $(HELM) $(KUSTOMIZE) $(VGOPATH) $(YQ)
	@REPO_ROOT=$(REPO_ROOT) VGOPATH=$(VGOPATH) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) bash $(GARDENER_HACK_DIR)/generate-sequential.sh ./charts/... ./example/... ./cmd/... ./pkg/...
	@REPO_ROOT=$(REPO_ROOT) VGOPATH=$(VGOPATH) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) $(REPO_ROOT)/hack/update-codegen.sh
	@bash ./hack/generate-renovate-ignore-deps.sh

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@bash $(GARDENER_HACK_DIR)/format.sh ./cmd ./pkg ./test

.PHONY: sast
sast: $(GOSEC)
	@bash $(GARDENER_HACK_DIR)/sast.sh --exclude-dirs hack,gardener

.PHONY: sast-report
sast-report: $(GOSEC)
	@bash $(GARDENER_HACK_DIR)/sast.sh --exclude-dirs hack,gardener --gosec-report true

.PHONY: test
test:
	@bash $(GARDENER_HACK_DIR)/test.sh ./cmd/... ./pkg/...

.PHONY: test-cov
test-cov:
	@bash $(GARDENER_HACK_DIR)/test-cover.sh ./cmd/... ./pkg/...

.PHONY: test-clean
test-clean:
	@bash $(GARDENER_HACK_DIR)/test-cover-clean.sh

.PHONY: test-integration
test-integration: $(REPORT_COLLECTOR) $(SETUP_ENVTEST)
	@bash $(GARDENER_HACK_DIR)/test-integration.sh ./test/integration/controller/...

.PHONY: verify
verify: check format test test-integration sast

.PHONY: verify-extended
verify-extended: check-generate check format test-cov test-clean sast-report
