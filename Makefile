BIN_DIR = ./bin
GOLANGCILINT_VERSION = 1.17.1

export GO111MODULE=on

.PHONY: all
all: lint test build

test:
	go test ./...

$(BIN_DIR)/golangci-lint: $(BIN_DIR)
	@wget -O - -q https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINDIR=$(@D) sh -s v$(GOLANGCILINT_VERSION) > /dev/null 2>&1

.PHONY: lint
lint: $(BIN_DIR)/golangci-lint
	$(BIN_DIR)/golangci-lint run -c .golangci.yml

.PHONY: build
build:
	go build -o bin/ratio cmd/server/main.go
