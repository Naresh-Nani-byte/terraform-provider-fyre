THIS_FILE := $(lastword $(MAKEFILE_LIST))

PROVIDER_HOSTNAME=registry.terraform.io
PROVIDER_NAMESPACE=hashicorp-forge
PROVIDER_NAME?=fyre
PROVIDER_BIN_NAME=terraform-provider-fyre
PROVIDER_BIN_OS?=$$(go env GOOS)
PROVIDER_BIN_ARCH?=$$(go env GOARCH)
PROVIDER_BIN_VERSION?=$$(cat VERSION)
PROVIDER_BUILD_TAGS?=-tags osusergo
PROVIDER_LD_FLAGS?=-ldflags="-extldflags=-static"
CI?=false
TEST?=./...
TEST_BLD_DIR=./test-build

SHELL = /bin/bash

.PHONY: default
default: build install

.PHONY: all
all: clean generate lint fmt-check install test-race-detector

.PHONY: build
build:
	mkdir -p ./dist
	CGO_ENABLED=0 GOOS=${PROVIDER_BIN_OS} GOARCH=${PROVIDER_BIN_ARCH} go build ${PROVIDER_BUILD_TAGS} ${PROVIDER_LD_FLAGS} -o ./dist/${PROVIDER_BIN_NAME}_${PROVIDER_BIN_VERSION}_${PROVIDER_BIN_OS}_${PROVIDER_BIN_ARCH}

.PHONY: build-race-detector
build-race-detector:
	CGO_ENABLED=0 GOOS=${PROVIDER_BIN_OS} GOARCH=${PROVIDER_BIN_ARCH} go build -race ${PROVIDER_BUILD_TAGS} ${PROVIDER_LD_FLAGS} -o ./dist/${PROVIDER_BIN_NAME}_${PROVIDER_BIN_VERSION}_${PROVIDER_BIN_OS}_${PROVIDER_BIN_ARCH}

.PHONY: docs
docs:
	type tfplugindocs || go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	tfplugindocs generate --examples-dir examples --provider-dir . --provider-name fyre --rendered-website-dir docs

.PHONY: check-doc-delta
check-doc-delta:
	rm -rf ./docs
	@$(MAKE) -f $(THIS_FILE) docs
	@if ! git diff --exit-code; then echo "Documentation need to be regenerated. Run 'make docs' to fix them." && exit 1; fi

.PHONY: generate
generate:
	pushd tools &&  go generate ./... && popd

.PHONY: install
install:
	for binary in $$(ls ./dist | grep ${PROVIDER_BIN_NAME}) ; do \
	version=$$(echo $$binary | cut -d "_" -f 2); \
	platform=$$(echo $$binary | cut -d "_" -f 3); \
	arch=$$(echo $$binary | cut -d "_" -f 4); \
	mkdir -p ~/.terraform.d/plugins/${PROVIDER_HOSTNAME}/${PROVIDER_NAMESPACE}/${PROVIDER_NAME}/$${version}/$${platform}_$${arch}; \
	cp ./dist/$$binary ~/.terraform.d/plugins/${PROVIDER_HOSTNAME}/${PROVIDER_NAMESPACE}/${PROVIDER_NAME}/$${version}/$${platform}_$${arch}/${PROVIDER_BIN_NAME}; \
	chmod +x ~/.terraform.d/plugins/${PROVIDER_HOSTNAME}/${PROVIDER_NAMESPACE}/${PROVIDER_NAME}/$${version}/$${platform}_$${arch}/*; \
done

.PHONY: test
test:
	go test $(TEST) -v $(TESTARGS) -timeout=5m -parallel=4

.PHONY: test-acc
test-acc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 12m

.PHONY: test-race-detector
test-race-detector:
	GORACE=log_path=/tmp/gorace.log TF_ACC=1 go test -race $(TEST) -v $(TESTARGS) -timeout 120m

.PHONY: fmt
fmt: fmt-golang fmt-enos

.PHONY: fmt-golang
fmt-golang:
	gofumpt -w -l .

.PHONY: fmt-enos
fmt-enos:
	enos fmt enos
	terraform fmt -recursive enos
	terraform fmt -recursive examples

.PHONY: fmt-check
fmt-check: fmt-check-golang fmt-check-enos

.PHONY: fmt-check-golang
fmt-check-golang:
	gofumpt -d -l .

.PHONY: fmt-check-enos
fmt-check-enos:
	enos fmt -cd enos
	terraform fmt -recursive -check enos
	terraform fmt -recursive -check examples

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: lint-fix
lint-fix:
	golangci-lint run -v --fix

.PHONY: clean
clean:
	rm -rf dist/* bin/* .terraform*
