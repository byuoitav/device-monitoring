# git
GIT_ORG=byuoitav
GIT_BRANCH=development

# go
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GO_VENDOR=gvt fetch 

# docker
DOCKER=docker
DOCKER_BUILD=$(DOCKER) build
DOCKER_LOGIN=$(DOCKER) login
DOCKER_PUSH=$(DOCKER) push
DOCKER_FILE=Dockerfile-development
DOCKER_FILE_ARM=Dockerfile-ARM

# general
DIR=$(dir $(realpath $(lastword $(MAKEFILE_LIST))))
NAME := $(shell basename "$(PWD)")

all: vendor test build

build:
	$(GOBUILD) -o $(NAME) -v

build-arm: 
	env GOOS=linux GOARCH=arm GOARM=5 $(GOBUILD) -o $(NAME)-arm -v

test: 
	$(GOTEST) -v -race $(go list ./... | grep -v /vendor/) 

clean: 
	$(GOCLEAN)

run: build
	./$(BINARY_NAME)

vendor: 
	# add all vendored files here, with the correct version
	# fix this :) 
	$(VENDOR) github.com/labstack/echo
	$(VENDOR) github.com/fatih/color


deploy: 
	echo "$(CURDIR)"

docker-x86: 
	$(DOCKER_BUILD) -f $(DOCKER_FILE) -t $(GIT_ORG)/$(NAME):$(GIT_BRANCH) .
#	$(DOCKER_LOGIN) -e $(DOCKER_EMAIL) -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD) no email?
	$(DOCKER_LOGIN) -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
	$(DOCKER_PUSH) $(GIT_ORG)/$(NAME):$(GIT_BRANCH)

docker-arm:

