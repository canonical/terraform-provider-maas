TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=ionutbalutoiu
NAME=maas
BINARY=terraform-provider-${NAME}
VERSION=0.1
OS_ARCH=darwin_arm64

default: install

.PHONY: build
build:
	mkdir -p ./bin
	go build -o ./bin/${BINARY}

.PHONY: release
release:
	mkdir -p ./dist
	GOOS=darwin GOARCH=amd64 go build -o ./dist/${BINARY}_${VERSION}_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -o ./dist/${BINARY}_${VERSION}_darwin_arm64
	GOOS=freebsd GOARCH=386 go build -o ./dist/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./dist/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./dist/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./dist/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./dist/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./dist/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./dist/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./dist/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./dist/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./dist/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./dist/${BINARY}_${VERSION}_windows_amd64

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ./bin/${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

.PHONY: test
test:
	go test $(TEST) -v $(TESTARGS) -timeout=5m -parallel=4

.PHONY: mockgen
mockgen:
	go install github.com/golang/mock/mockgen@v1.5.0
	mockgen -destination ./test/mocks/gomaasapi.go github.com/juju/gomaasapi Machine,Controller
