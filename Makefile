GO_BIN_PATH = bin
GO_SRC_PATH = simple_tcp
GO_HTTP_SRC_PATH = http
ROOT_PATH = $(shell pwd)

NF = $(GO_NF)
GO_NF = server client NF_A
HTTP_NF = http_server http_client

NF_GO_FILES = $(shell find $(GO_SRC_PATH)/$(%) -name "*.go" ! -name "*_test.go")
HTTP_GO_FILES = $(shell find $(GO_HTTP_SRC_PATH)/$(%) -name "*.go" ! -name "*_test.go")

VERSION = $(shell git describe --tags)
BUILD_TIME = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH = $(shell git submodule status | grep $(GO_SRC_PATH)/$(@F) | awk '{print $$(1)}' | cut -c1-8)
COMMIT_TIME = $(shell cd $(GO_SRC_PATH)/$(@F) && git log --pretty="%ai" -1 | awk '{time=$$(1)"T"$$(2)"Z"; print time}')
LDFLAGS = -X github.com/free5gc/version.VERSION=$(VERSION) \
          -X github.com/free5gc/version.BUILD_TIME=$(BUILD_TIME) \
          -X github.com/free5gc/version.COMMIT_HASH=$(COMMIT_HASH) \
          -X github.com/free5gc/version.COMMIT_TIME=$(COMMIT_TIME)

.PHONY: $(NF) $(WEBCONSOLE) clean

.DEFAULT_GOAL: nfs

nfs: $(NF) http

all: $(NF)

$(GO_NF): % : $(GO_BIN_PATH)/%
$(HTTP_NF): % : $(GO_BIN_PATH)/%

$(GO_BIN_PATH)/%: %.go $(NF_GO_FILES)
# $(@F): The file-within-directory part of the file name of the target.
	@echo "Start building $(@F)...."
	cd $(GO_SRC_PATH)/$(@F) && \
	CGO_LDFLAGS_ALLOW="-Wl,(--whole-archive|--no-whole-archive)" go build -ldflags "$(LDFLAGS)" -o $(ROOT_PATH)/$@ $(@F).go

vpath %.go $(addprefix $(GO_SRC_PATH)/, $(GO_NF))

http: $(HTTP_NF)
$(GO_BIN_PATH)/%: $(HTTP_GO_FILES)
	@echo "Start building $(@F)...."
	cd $(GO_HTTP_SRC_PATH)/$(@F) && \
	CGO_LDFLAGS_ALLOW="-Wl,(--whole-archive|--no-whole-archive)" go build -ldflags "$(LDFLAGS)" -o $(ROOT_PATH)/$@ $(@F).go
