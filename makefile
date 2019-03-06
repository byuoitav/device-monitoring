# vars
ORG=$(shell echo $(CIRCLE_PROJECT_USERNAME))
BRANCH=$(shell echo $(CIRCLE_BRANCH))
NAME=$(shell echo $(CIRCLE_PROJECT_REPONAME))

ifeq ($(NAME),)
NAME := $(shell basename "$(PWD)")
endif

ifeq ($(ORG),)
ORG=byuoitav
endif

ifeq ($(BRANCH),)
BRANCH:= $(shell git rev-parse --abbrev-ref HEAD)
endif

# go
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
VENDOR=gvt fetch -branch $(BRANCH)

# angular
NPM=npm
NPM_INSTALL=$(NPM) install
NPM_BUILD=npm run-script build
NG1=dashboard

# aws
AWS_S3_ADD=aws s3 cp
S3_BUCKET=$(shell echo $(AWS_S3_SERVICES_BUCKET))

all: deploy clean

ci: deps all

build: build-x86 build-arm build-web

build-x86:
	env GOOS=linux CGO_ENABLED=0 $(GOBUILD) -o $(NAME)-bin -v

build-arm:
	env GOOS=linux GOARCH=arm $(GOBUILD) -o $(NAME)-arm -v

build-web: $(NG1)
	cd $(NG1) && $(NPM_INSTALL) && $(NPM_BUILD)
	mkdir files
	mv $(NG1)/dist/$(NG1) files/$(NG1)

test:
	$(GOTEST) -v -race $(go list ./... | grep -v /vendor/)

clean:
	$(GOCLEAN)
	rm -f $(NAME)-bin
	rm -f $(NAME)-arm
	rm -rf files/

run: $(NAME)-bin
	./$(NAME)-bin

deps:
	# TODO remove whenever this npm bug is fixed
	# https://github.com/npm/npm/issues/20861
	npm config set unsafe-perm true
	$(NPM_INSTALL) -g @angular/cli
ifneq "$(BRANCH)" "master"
	# put vendored packages in here
	# e.g. $(VENDOR) github.com/byuoitav/event-router-microservice
	gvt fetch -tag v3.3.10 github.com/labstack/echo
	$(VENDOR) github.com/byuoitav/common
	$(VENDOR) github.com/byuoitav/central-event-system
endif
	$(GOGET) -d -v

deploy: $(NAME)-arm $(NAME).service.tmpl files/$(NG1) version.txt
ifeq "$(BRANCH)" "master"
	$(eval BRANCH=development)
endif
	@echo adding files to $(S3_BUCKET)
	$(AWS_S3_ADD) $(NAME)-arm s3://$(S3_BUCKET)/$(BRANCH)/$(NAME)/$(NAME)
	$(AWS_S3_ADD) $(NAME).service.tmpl s3://$(S3_BUCKET)/$(BRANCH)/$(NAME)/device-monitoring.service.tmpl
	$(AWS_S3_ADD) version.txt s3://$(S3_BUCKET)/$(BRANCH)/$(NAME)/files/version.txt
	$(AWS_S3_ADD) files/ s3://$(S3_BUCKET)/$(BRANCH)/$(NAME)/files/ --recursive
ifeq "$(BRANCH)" "development"
	$(eval BRANCH=master)
endif

### deps
$(NAME)-bin:
	$(MAKE) build-x86

$(NAME)-arm:
	$(MAKE) build-arm

files/$(NG1):
	$(MAKE) build-web
