# =============================
# Build & Deploy Only
# =============================
SHELL := /bin/bash

# Metadata
ORG    ?= $(shell echo $(CIRCLE_PROJECT_USERNAME))
NAME   ?= $(shell echo $(CIRCLE_PROJECT_REPONAME))
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
ifeq ($(NAME),)
NAME := $(shell basename "$(PWD)")
endif
ifeq ($(ORG),)
ORG := byuoitav
endif


# Tools
GOCMD   = go
GOBUILD = $(GOCMD) build
NPM     = npm
NG1     = dashboard

# Outputs
BUILD_DIR  = dist
BIN_OUTPUT = $(BUILD_DIR)/$(NAME)

# Cross-build matrix for build-binaries
PLATFORMS = linux/amd64 linux/arm


# Go build flags (all optional, override on make cmdline)
GOOS        ?=
GOARCH      ?=
GOARM       ?=
CGO_ENABLED ?=
TAGS        ?=
LDFLAGS     ?=
GCFLAGS     ?=
ASMFLAGS    ?=
BUILD_FLAGS ?=
MAIN_PKG    ?= .

# Compose go build flags
COMMON_BUILD_FLAGS :=
ifneq ($(strip $(TAGS)),)
  COMMON_BUILD_FLAGS += -tags '$(TAGS)'
endif
ifneq ($(strip $(LDFLAGS)),)
  COMMON_BUILD_FLAGS += -ldflags '$(LDFLAGS)'
endif
ifneq ($(strip $(GCFLAGS)),)
  COMMON_BUILD_FLAGS += -gcflags '$(GCFLAGS)'
endif
ifneq ($(strip $(ASMFLAGS)),)
  COMMON_BUILD_FLAGS += -asmflags '$(ASMFLAGS)'
endif
ifneq ($(strip $(BUILD_FLAGS)),)
  COMMON_BUILD_FLAGS += $(BUILD_FLAGS)
endif

# =============================
# Targets
# =============================
.PHONY: all build-local build-binaries build-web clean deploy

all: build-web build-local

ci: deps all test

# Default local build targets linux/arm (GOARM=7)
build-local:
	@echo "Building $(NAME) for linux/arm (GOARM=7) — override with GOOS/GOARCH/GOARM/CGO_ENABLED as needed..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=$${GOOS:-linux} \
	 GOARCH=$${GOARCH:-arm} \
	 GOARM=$${GOARM:-7} \
	 CGO_ENABLED=$${CGO_ENABLED:-0} \
	 $(GOBUILD) $(COMMON_BUILD_FLAGS) -o $(BIN_OUTPUT) -v $(MAIN_PKG)

build-binaries:
	@echo "Building binaries for: $(PLATFORMS)"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; ARCH=$${platform#*/}; \
		OUT=$(BUILD_DIR)/$(NAME)-$$OS-$$ARCH; \
		[ "$$OS" = "windows" ] && OUT=$$OUT.exe; \
		echo "  -> $$OS/$$ARCH => $$OUT"; \
		if [ "$$ARCH" = "arm" ]; then \
		  GOOS=$$OS GOARCH=$$ARCH GOARM=$${GOARM:-7} CGO_ENABLED=$${CGO_ENABLED:-0} \
		    $(GOBUILD) $(COMMON_BUILD_FLAGS) -o "$$OUT" -v $(MAIN_PKG) || exit 1; \
		else \
		  GOOS=$$OS GOARCH=$$ARCH CGO_ENABLED=$${CGO_ENABLED:-0} \
		    $(GOBUILD) $(COMMON_BUILD_FLAGS) -o "$$OUT" -v $(MAIN_PKG) || exit 1; \
		fi; \
	done

build-web:
	@echo "Building Angular dashboard..."
	. $$HOME/.nvm/nvm.sh && nvm use 20.19.0 \
	  && cd dashboard \
	  && $(NPM) install \
	  && ./node_modules/.bin/ng build --configuration production --base-href /dashboard/
	@echo "Copying built web assets..."
	mkdir -p files/$(NG1)
	@rsync -a --delete dashboard/dist/$(NG1)/ files/$(NG1)/

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) files vendor *.tar.gz dist

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
	@echo "Packaging $(BRANCH).tar.gz"
	@cp version.txt files/
	@cp service-config.json files/
	@tar -czf $(BRANCH).tar.gz \
	  -C $(CURDIR) files \
	  -C $(BUILD_DIR) $(NAME)
	@echo "Fetching current CouchDB doc revision…"
	$(eval rev=$(shell curl -s -n -X GET -u ${DB_USERNAME}:${DB_PASSWORD} "${DB_ADDRESS}/deployment-information/$(NAME)" | cut -d, -f2 | cut -d\" -f4))
	@echo "Uploading tarball to CouchDB…"
	@curl -X PUT -u ${DB_USERNAME}:${DB_PASSWORD} \
	  -H "Content-Type: application/gzip" \
	  -H "If-Match: $(rev)" \
	  ${DB_ADDRESS}/deployment-information/$(NAME)/$(BRANCH).tar.gz \
	  --data-binary @$(BRANCH).tar.gz
ifeq ($(BRANCH),development)
	$(eval BRANCH=master)
endif

# Build triggers
$(BIN_OUTPUT):
	$(MAKE) build-local

files/$(NG1):
	$(MAKE) build-web
