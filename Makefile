REPO := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

ifdef GOROOT
	GO = $(GOROOT)/bin/go
else
	GO = $(shell which go)
	GOROOT = $(shell $(GO) env GOROOT)
endif

all: build

build:
	$(GO) install -v github.com/majestrate/bdsmail/cmd/maild
	$(GO) install -v github.com/majestrate/bdsmail/cmd/mailtool
	$(GO) install -v github.com/majestrate/bdsmail/cmd/bdsconfig

clean:
	$(GO) clean -v
	rm -fr $(REPO)/pkg
test:
	$(GO) test github.com/majestrate/bdsmail/lib/...
