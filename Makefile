TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=maas
NAME=maas
BINARY=terraform-provider-${NAME}
VERSION=1.0.1

OS?=$$(go env GOOS)
ARCH?=$$(go env GOARCH)

TEST_PARALLELISM?=4

default: install

BIN=$(CURDIR)/bin
$(BIN)/%:
	@echo "Installing tools from tools/tools.go"
	@cat tools/tools.go | grep _ | awk -F '"' '{print $$2}' | GOBIN=$(BIN) xargs -tI {} go install {}

.PHONY: build install clean clean_install test testacc generate_docs validate_docs tfproviderlintx tfproviderlint

build:
	mkdir -p $(BIN)
	go build -o $(BIN)/${BINARY}
	
create_dev_overrides: build
	@sh -c "'$(CURDIR)/scripts/generate-dev-overrides.sh'"

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS}_${ARCH}
	mv $(BIN)/${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS}_${ARCH}

clean:
	rm -rf $(BIN)

clean_install: clean
	rm -rf ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}

test:
	go test $(TEST) -v $(TESTARGS) -timeout=5m -parallel=$(TEST_PARALLELISM)

testacc:
# TODO: do these values also need to be added here
# MAAS_API_VERSION=2.0
# MAAS_API_KEY=// TODO this will need to be dynamically allocated
# MAAS_API_URL=http://127.0.0.1:5240/MAAS
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m -parallel=$(TEST_PARALLELISM)

generate_docs: $(BIN)/tfplugindocs
	$(BIN)/tfplugindocs generate --provider-name $(NAME)

validate_docs: $(BIN)/tfplugindocs
	$(BIN)/tfplugindocs validate


tfproviderlintx: $(BIN)/tfproviderlint
	$(BIN)/tfproviderlintx $(TFPROVIDERLINT_ARGS) ./...

tfproviderlint: $(BIN)/tfproviderlintx
	$(BIN)/tfproviderlint $(TFPROVIDERLINT_ARGS) ./...