# go params
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
VENDOR=gvt fetch 
#DIR=$(dir $(realpath $(lastword $(MAKEFILE_LIST))))
#BINARY_NAME=binary

all: vendor test build

build:
	$(GOBUILD) -v

test: 
	$(GOTEST) -v -race $(go list ./... | grep -v /vendor/) 

clean: 
	$(GOCLEAN)

run: build
	./$(BINARY_NAME)

vendor: 
	# add all vendored files here, with the correct version
	$(VENDOR) github.com/labstack/echo
	$(VENDOR) github.com/fatih/color
