BINARY = usg-dns-api
SOURCE_FILES = $(shell find . -type f -name '*.go' -not -name '*_test.go')

MAIN_PKG    = $(shell go list)
VERSION_PKG = ${MAIN_PKG}/version
CMD_PKG     = ${MAIN_PKG}/cmd
SERVER_PKG  = ${MAIN_PKG}/server
DB_PKG      = ${MAIN_PKG}/db

DEFAULT_CONF_FILE  ?= usg-dns-api.yaml
DEFAULT_DB_FILE    ?= usg-dns-api.db
DEFAULT_HOSTS_FILE ?= hosts

VERSION    ?= $(shell ./generate-version.sh)
LAST_COMMIT = $(shell git rev-parse HEAD)

DIST_DIR = ./dist

TEST_LOCATION ?= ./...
TEST_CMD       = go test -v -race -cover

LD_FLAGS = -ldflags "-w -s -X ${VERSION_PKG}.commit=${LAST_COMMIT} -X ${VERSION_PKG}.version=${VERSION} -X ${CMD_PKG}.defaultConfigFile=${DEFAULT_CONF_FILE} -X ${DB_PKG}.defaultPath=${DEFAULT_DB_FILE} -X ${SERVER_PKG}.defaultHostsFile=${DEFAULT_HOSTS_FILE}"

all: $(BINARY)-$(shell go env GOOS)-$(shell go env GOARCH)

compile: $(BINARY)

$(BINARY): $(BINARY)-linux-amd64 $(BINARY)-linux-mips64 $(BINARY)-linux-arm64

$(BINARY)-linux-amd64: $(DIST_DIR)/$(BINARY)-linux-amd64

$(BINARY)-linux-mips64: $(DIST_DIR)/$(BINARY)-linux-mips64

$(BINARY)-linux-arm64: $(DIST_DIR)/$(BINARY)-linux-arm64

$(DIST_DIR)/$(BINARY)-linux-amd64: $(SOURCE_FILES) go.mod
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LD_FLAGS) -o $@ ./main.go

$(DIST_DIR)/$(BINARY)-linux-mips64: $(SOURCE_FILES) go.mod
	CGO_ENABLED=0 GOOS=linux GOARCH=mips64 go build $(LD_FLAGS) -o $@ ./main.go

$(DIST_DIR)/$(BINARY)-linux-arm64: $(SOURCE_FILES) go.mod
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LD_FLAGS) -o $@ ./main.go

.PHONY: test
test:
	$(TEST_CMD) $(COVER_OPTS) $(TEST_LOCATION)

.PHONY: clean
clean:
	rm -f $(DIST_DIR)/$(BINARY)-amd64 $(DIST_DIR)/$(BINARY)-mips64
