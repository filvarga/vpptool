GOCMD=go
GOINSTALL=$(GOCMD) install
GOBUILD=$(GOCMD) build

TOOL_NAME=./cmd/vpptool

install:
	$(GOINSTALL) $(TOOL_NAME)

build:
	$(GOBUILD) $(TOOL_NAME)
