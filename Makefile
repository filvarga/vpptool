# vim: set ts=2:
GOCMD=go
GOINSTALL=$(GOCMD) install
GOBUILD=$(GOCMD) build

BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

TOOL_NAME=vpptool

CONTEXT = "https://github.com/filvarga/vpptool.git\#$(BRANCH):docker/vpptool"

LDFLAGS=-ldflags "-X main.context=$(CONTEXT)" 

all: install

vpptool:
	$(GOBUILD) $(LDFLAGS) -o build/$(TOOL_NAME) ./cmd/$(TOOL_NAME)

install:
	$(GOINSTALL) $(LDFLAGS) ./cmd/$(TOOL_NAME)
