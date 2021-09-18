.PHONY: build clean install start help

PROJECT := "xybwcollect"
GO_MOD := github.com/fufuok/xy-bw-collect

GO_BIN := /usr/local/go117/bin/go

CI_BIN := ../bin/$(PROJECT)
CI_ETC := ../etc/$(PROJECT).json
OPT_DIR := /opt/xunyou/$(PROJECT)
OPT_BIN := $(OPT_DIR)/bin
OPT_LOG := $(OPT_DIR)/log
OPT_ETC := $(OPT_DIR)/etc

GIT_TAG := $(shell git describe --tags --abbrev=0)
GIT_COMMIT := $(shell git log $(GIT_TAG) --pretty=format:"%an-[%h] * %s" -1)
ifeq ($(GIT_TAG),)
	VERSION=dev
else
	VERSION=$(GIT_TAG)
endif
BUILD_TIME := $(shell date +%y%m%d%H%M%S)
GO_VERSION := $(shell $(GO_BIN) version)
LDFLAGS = -X '$(GO_MOD)/conf.Version=$(VERSION).$(BUILD_TIME)' \
-X '$(GO_MOD)/conf.GitCommit=$(GIT_COMMIT)' \
-X '$(GO_MOD)/conf.GoVersion=$(GO_VERSION)'

all: build

build:
	$(GO_BIN) build -tags=go_json -ldflags "$(LDFLAGS)" -v -o $(CI_BIN) cli/$(PROJECT)/main.go

clean:
	rm -f $(CI_BIN)

install:
	mkdir -p $(DESTDIR)/$(OPT_DIR)/bin
	mkdir -p $(DESTDIR)/$(OPT_DIR)/etc
	mkdir -p $(DESTDIR)/$(OPT_DIR)/log
	install -m 755 -D $(CI_BIN) $(DESTDIR)/$(OPT_DIR)/bin/$(PROJECT)
	install -m 644 -D $(CI_ETC) $(DESTDIR)/$(OPT_DIR)/etc/$(PROJECT).json

start:
	$(OPT_DIR)/bin/$(PROJECT)

help:
	@echo make: compile packages and dependencies
	@echo make clean: stop service and remove object files
	@echo make install: deploy
	@echo make start: start service
