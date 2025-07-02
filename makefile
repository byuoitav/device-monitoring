# =============================
# Metadata and Project Defaults
# =============================

ORG     ?= $(shell echo $(CIRCLE_PROJECT_USERNAME))
NAME    ?= $(shell echo $(CIRCLE_PROJECT_REPONAME))
BRANCH  ?= $(shell git rev-parse --abbrev-ref HEAD)

ifeq ($(NAME),)
NAME := $(shell basename "$(PWD)")
endif

ifeq ($(ORG),)
ORG := byuoitav
endif

# =============================
# Tools and Commands
# =============================

GOCMD      = go
GOBUILD    = $(GOCMD) build
GOCLEAN    = $(GOCMD) clean
GOTEST     = $(GOCMD) test
GOGET      = $(GOCMD) get
VENDOR     = gvt fetch -branch $(BRANCH)

NPM        = npm
NPM_INSTALL= $(NPM) install
NPM_BUILD  = npm run-script build
NG1        = dashboard

BUILD_DIR  = dist
BIN_OUTPUT = $(BUILD_DIR)/$(NAME)

PLATFORMS  = linux/amd64 linux/arm

# =============================
# Main Targets
# =============================

all: build-web build-local

ci: deps all test

build-local:
	$(GOBUILD) -o $(BIN_OUTPUT) -v

build-binaries:
	@echo "Building binaries for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; ARCH=$${platform#*/}; \
		OUTPUT=$(BUILD_DIR)/$(NAME)-$$OS-$$ARCH; \
		[ "$$OS" = "windows" ] && OUTPUT=$$OUTPUT.exe; \
		echo "Building $$OS/$$ARCH -> $$OUTPUT"; \
		GOOS=$$OS GOARCH=$$ARCH $(GOBUILD) -o $$OUTPUT -v || exit 1; \
	done

build-web: $(NG1)
	cd $(NG1) && $(NPM_INSTALL) && $(NPM_BUILD)
	mkdir -p files
	mv $(NG1)/dist/$(NG1) files/$(NG1)

test:
	$(GOTEST) -v -race $$(go list ./... | grep -v /vendor/)

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(NAME)
	rm -rf $(BUILD_DIR) files vendor
	rm -f *.tar.gz

deps:
	npm config set unsafe-perm true
	$(NPM_INSTALL) -g @angular/cli
ifneq ($(BRANCH),master)
	$(VENDOR) github.com/labstack/echo@v3.3.10
	$(VENDOR) github.com/go-redis/redis@v6.15.3
	$(VENDOR) github.com/byuoitav/common
	$(VENDOR) github.com/byuoitav/central-event-system
	$(VENDOR) github.com/byuoitav/shipwright
endif
	$(GOGET) -d -v

# =============================
# Deployment
# =============================

deploy: $(BIN_OUTPUT) files/$(NG1) version.txt
ifeq ($(BRANCH),master)
	$(eval BRANCH=development)
endif
	@echo Building deployment tarball
	@cp version.txt files/
	@cp service-config.json files/
	$(eval BRANCH_FILENAME := $(subst /,-,$(BRANCH)))
	@tar -czf $(BRANCH_FILENAME).tar.gz $(BIN_OUTPUT) files
	@echo Getting current doc revision
	$(eval rev=$(shell curl -s -n -X GET -u ${DB_USERNAME}:${DB_PASSWORD} "${DB_ADDRESS}/deployment-information/$(NAME)" | cut -d, -f2 | cut -d\" -f4))
	@echo Pushing zip up to couch
	@curl -X PUT -u ${DB_USERNAME}:${DB_PASSWORD} -H "Content-Type: application/gzip" -H "If-Match: $(rev)" ${DB_ADDRESS}/deployment-information/$(NAME)/$(BRANCH).tar.gz --data-binary @$(BRANCH_FILENAME).tar.gz
ifeq ($(BRANCH),development)
	$(eval BRANCH=master)
endif

# Build triggers
$(BIN_OUTPUT): 
	$(MAKE) build-local

files/$(NG1):
	$(MAKE) build-web

# Debug helper
print-vars:
	@echo "NAME=$(NAME)"
	@echo "BRANCH=$(BRANCH)"
	@echo "BIN_OUTPUT=$(BIN_OUTPUT)"
