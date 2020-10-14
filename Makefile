# vim: set ts=2:
GOCMD=go
GOINSTALL=$(GOCMD) install
GOBUILD=$(GOCMD) build

TOOL_NAME=vpptool

all: install

vpptool:
	$(GOBUILD) -o build/$(TOOL_NAME) ./cmd/$(TOOL_NAME)

install:
	$(GOINSTALL) ./cmd/$(TOOL_NAME)
