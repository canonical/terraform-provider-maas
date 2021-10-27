TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=ionutbalutoiu
NAME=maas
BINARY=terraform-provider-${NAME}
VERSION=1.0.1
OS?=$$(go env GOOS)
ARCH?=$$(go env GOARCH)

default: install

.PHONY: build
build:
	mkdir -p ./bin
	go build -o ./bin/${BINARY}

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS}_${ARCH}
	mv ./bin/${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS}_${ARCH}

.PHONY: test
test:
	go test $(TEST) -v $(TESTARGS) -timeout=5m -parallel=4
