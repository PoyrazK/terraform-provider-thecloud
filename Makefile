HOSTNAME=github.com
NAMESPACE=poyrazk
NAME=thecloud
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
OS_ARCH=darwin_arm64

default: build

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: build install test testacc
