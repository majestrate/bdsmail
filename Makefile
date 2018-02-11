REPO := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

ifdef GOROOT
	GO = $(GOROOT)/bin/go
else
	GO = $(shell which go)
endif


all: build

build:
	GOPATH=$(REPO) $(GO) install -v bds/cmd/maild
	GOPATH=$(REPO) $(GO) install -v bds/cmd/newmail
	GOPATH=$(REPO) $(GO) install -v bds/cmd/bdsconfig

clean:
	go clean -v
	rm -rf pkg

test:
	GOPATH=$(REPO) go test bds/lib/...
