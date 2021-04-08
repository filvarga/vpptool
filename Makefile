# vim: set ts=2:
GOCMD=go
GOINSTALL=$(GOCMD) install
GOBUILD=$(GOCMD) build

BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

TOOL_NAME=vpptool

CONTEXT = "https://github.com/filvarga/vpptool.git\#$(BRANCH):docker"
GIT_URL = "https://github.com/FDio/vpp.git"
GO_VERSION=1.16.3
CS_VERSION=3.9.2

LDFLAGS=-ldflags "-X main.context=$(CONTEXT) -X main.git_url=$(GIT_URL) -X main.cs_version=$(CS_VERSION) -X main.go_version=$(GO_VERSION)" 

all: install

vpptool:
	$(GOBUILD) $(LDFLAGS) -o build/$(TOOL_NAME) ./cmd/$(TOOL_NAME)

install:
	$(GOINSTALL) $(LDFLAGS) ./cmd/$(TOOL_NAME)
