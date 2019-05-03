# Parameters
PKG = github.com/logentries_exporter
NAME = logentries_exporter
DOC = README.md LICENSE

# Replace backslashes with forward slashes for use on Windows.
# Make is !@#$ing weird.
E :=
BSLASH := \$E
FSLASH := /

# Directories
WD := $(subst $(BSLASH),$(FSLASH),$(shell pwd))
MD := $(subst $(BSLASH),$(FSLASH),$(shell dirname "$(realpath $(lastword $(MAKEFILE_LIST)))"))
PKGDIR = $(MD)
CMDDIR = $(PKGDIR)
DISTDIR ?= $(WD)/dist
RPTDIR ?= $(WD)/reports
GP = $(subst $(BSLASH),$(FSLASH),$(GOPATH))

# Parameters
VERSION ?= $(shell git -C "$(MD)" describe --abbrev=0 --tags)
COMMIT_ID := $(shell git -C "$(MD)" rev-parse HEAD | head -c8)
BUILD_TAGS ?= release
CMDPKG = $(PKG)
CMDS := $(shell find "$(CMDDIR)/" -mindepth 1 -maxdepth 1 -type d | sed 's/ /\\ /g' | xargs -n1 basename)
BENCHCPUS ?= 1,2,4

# Commands
GOCMD = go
ARCHES ?= 386 amd64
OSES ?= linux darwin
OUTTPL = $(DISTDIR)/$(NAME)-$(VERSION)-{{.OS}}_{{.Arch}}/{{.Dir}}
LDFLAGS = -X $(PKG).Version=$(VERSION) -X $(PKG).Build=$(COMMIT_ID)
GOBUILD = gox -rebuild -gocmd="$(GOCMD)" -arch="$(ARCHES)" -os="$(OSES)" -output="$(OUTTPL)" -tags "$(BUILD_TAGS)" -ldflags "$(LDFLAGS)"
GOCLEAN = $(GOCMD) clean
GOINSTALL = $(GOCMD) install -a -tags "$(BUILD_TAGS)" -ldflags "$(LDFLAGS)"
GOTEST = $(GOCMD) test -v -tags "$(BUILD_TAGS)"
GOLINT = gometalinter --deadline=30s --tests --disable=aligncheck --disable=gocyclo --disable=gotype
GODEP = $(GOCMD) get -d -t
GOFMT = goreturns -w
GOBENCH = $(GOCMD) test -v -tags "$(BUILD_TAGS)" -cpu=$(BENCHCPUS) -run=NOTHING -bench=. -benchmem -outputdir "$(RPTDIR)"
GZCMD = tar -czf
ZIPCMD = zip
SHACMD = sha256sum
SLOCCMD = cloc --by-file --xml --exclude-dir="vendor" --include-lang="Go"
XUCMD = go2xunit

# Dynamic Targets
INSTALL_TARGETS := $(addprefix install-,$(CMDS))

.PHONY: all

all: debug setup dep format lint test bench build dist
release: clean debug setup dep format build dist

setup: setup-dirs setup-build setup-format setup-lint setup-reports
setup-reports: setup-dirs
	go get github.com/tebeka/go2xunit
setup-build: setup-dirs
	go get github.com/mitchellh/gox
setup-format: setup-dirs
	go get github.com/sqs/goreturns
setup-lint: setup-dirs
	go get github.com/alecthomas/gometalinter
	gometalinter --install
setup-dirs:
	mkdir -p "$(RPTDIR)"
	mkdir -p "$(DISTDIR)"
clean:
	$(GOCLEAN) $(PKG)
	rm -rf "$(DISTDIR)"/*
	rm -f "$(RPTDIR)"/*
format:
	$(GOFMT) "$(PKGDIR)"
dep:
	$(GODEP) $(PKG)/...
lint: setup-dirs dep
	$(GOLINT) "$(PKGDIR)" | tee "$(RPTDIR)/lint.out"
check: setup-dirs clean dep
	$(GOTEST) $$(go list "$(PKG)/..." | grep -v /vendor/) | tee "$(RPTDIR)/test.out"
bench: setup-dirs clean dep
	$(GOBENCH) $$(go list "$(PKG)/..." | grep -v /vendor/) | tee "$(RPTDIR)/bench.out"
report: check
	cd "$(PKGDIR)";$(SLOCCMD) --out="$(RPTDIR)/cloc.xml" . | tee "$(RPTDIR)/cloc.out"
	cat "$(RPTDIR)/test.out" | $(XUCMD) -output "$(RPTDIR)/tests.xml"
	go list -f '{{join .Deps "\n"}}' "$(CMDPKG)/..." | sort | uniq | xargs -I {} sh -c "go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' {} | tee -a '$(RPTDIR)/deps.out'"
build:
	$(GOBUILD)
install: $(INSTALL_TARGETS)
$(INSTALL_TARGETS):
	$(GOINSTALL) "$(CMDPKG)/$(subst install-,,$@)"

dist: build
	for docfile in $(DOC); do \
		for dir in "$(DISTDIR)"/*; do \
			cp "$(PKGDIR)/$$docfile" "$$dir/"; \
		done; \
	done
	cd "$(DISTDIR)"; for dir in ./*linux*; do $(GZCMD) "$(basename "$$dir").tar.gz" "$$dir"; done
	cd "$(DISTDIR)"; for dir in ./*darwin*; do $(GZCMD) "$(basename "$$dir").tar.gz" "$$dir"; done
	cd "$(DISTDIR)"; find . -maxdepth 1 -type f -printf "$(SHACMD) %P | tee \"./%P.sha\"\n" | sh
	$(info "Built $(VERSION), build $(COMMIT_ID)")
debug:
	$(info MD=$(MD))
	$(info WD=$(WD))
	$(info PKG=$(PKG))
	$(info PKGDIR=$(PKGDIR))
	$(info DISTDIR=$(DISTDIR))
	$(info VERSION=$(VERSION))
	$(info COMMIT_ID=$(COMMIT_ID))
	$(info BUILD_TAGS=$(BUILD_TAGS))
	$(info CMDS=$(CMDS))
	$(info BUILD_TARGETS=$(BUILD_TARGETS))
	$(info INSTALL_TARGETS=$(INSTALL_TARGETS))
