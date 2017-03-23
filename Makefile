ROOT_PACKAGE := github.com/travis-ci/vsphere-monitor
MAIN_PACKAGE := $(ROOT_PACKAGE)/cmd/vsphere-monitor

VERSION_VAR := main.VersionString
VERSION_VALUE ?= $(shell git describe --always --dirty --tags 2>/dev/null)
REV_VAR := main.RevisionString
REV_VALUE ?= $(shell git rev-parse HEAD 2>/dev/null || echo "???")
REV_URL_VAR := main.RevisionURLString
REV_URL_VALUE ?= https://github.com/travis-ci/vsphere-monitor/tree/$(shell git rev-parse HEAD 2>/dev/null || echo "???")
GENERATED_VAR := main.GeneratedString
GENERATED_VALUE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%S%z')
COPYRIGHT_VAR := main.CopyrightString
COPYRIGHT_VALUE ?= $(shell grep -i ^copyright LICENSE | sed 's/^[Cc]opyright //')

GOPATH := $(shell go env GOPATH)
GOBUILD_LDFLAGS ?= \
	-X '$(VERSION_VAR)=$(VERSION_VALUE)' \
	-X '$(REV_VAR)=$(REV_VALUE)' \
	-X '$(REV_URL_VAR)=$(REV_URL_VALUE)' \
	-X '$(GENERATED_VAR)=$(GENERATED_VALUE)' \
	-X '$(COPYRIGHT_VAR)=$(COPYRIGHT_VALUE)'

.PHONY: all
all: clean build

.PHONY: clean
clean:
	go clean -i $(MAIN_PACKAGE)
	$(RM) -rv ./build

.PHONY: build
build: deps
	go install -ldflags "$(GOBUILD_LDFLAGS)" $(MAIN_PACKAGE)

.PHONY: deps
deps: vendor/.deps-fetched

vendor/.deps-fetched: vendor/manifest
	gvt rebuild
	touch $@
