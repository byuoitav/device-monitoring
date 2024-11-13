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

all: deploy clean

ci: deps all

build: build-web
	env GOOS=linux GOARCH=arm $(GOBUILD) -o $(NAME) -v

build-web: $(NG1)
	cd $(NG1) && $(NPM_INSTALL) && $(NPM_BUILD)
	mkdir files
	mv $(NG1)/dist/$(NG1) files/$(NG1)

test:
	$(GOTEST) -v -race $(go list ./... | grep -v /vendor/)

clean:
ifeq "$(BRANCH)" "master"
	$(eval BRANCH=development)
endif
	$(GOCLEAN)
	rm -f $(NAME)
	rm -f $(BRANCH).tar.gz
	rm -rf files/
	rm -rf vendor/
ifeq "$(BRANCH)" "development"
	$(eval BRANCH=master)
endif

deps:
	# TODO remove whenever this npm bug is fixed
	# https://github.com/npm/npm/issues/20861
	npm config set unsafe-perm true
	$(NPM_INSTALL) -g @angular/cli
ifneq "$(BRANCH)" "master"
	# put vendored packages in here
	# e.g. $(VENDOR) github.com/byuoitav/event-router-microservice
	gvt fetch -tag v3.3.10 github.com/labstack/echo
	gvt fetch -tag v6.15.3 github.com/go-redis/redis
	$(VENDOR) github.com/byuoitav/common
	$(VENDOR) github.com/byuoitav/central-event-system
	$(VENDOR) github.com/byuoitav/shipwright
endif
	$(GOGET) -d -v

deploy: $(NAME) files/$(NG1) version.txt
ifeq "$(BRANCH)" "master"
	$(eval BRANCH=development)
endif
	@echo Building deployment tarball
	@cp version.txt files/
	@cp service-config.json files/

	@tar -czf $(BRANCH).tar.gz $(NAME) files

	@echo Getting current doc revision
	$(eval rev=$(shell curl -s -n -X GET -u ${DB_USERNAME}:${DB_PASSWORD} "${DB_ADDRESS}/deployment-information/$(NAME)" | cut -d, -f2 | cut -d\" -f4))

	@echo Pushing zip up to couch
	@curl -X PUT -u ${DB_USERNAME}:${DB_PASSWORD} -H "Content-Type: application/gzip" -H "If-Match: $(rev)" ${DB_ADDRESS}/deployment-information/$(NAME)/$(BRANCH).tar.gz --data-binary @$(BRANCH).tar.gz
ifeq "$(BRANCH)" "development"
	$(eval BRANCH=master)
endif

### depsd
$(NAME):
	$(MAKE) build

files/$(NG1):
	$(MAKE) build-web
