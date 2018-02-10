

all: build

build:
	GOPATH=$(PWD) go install -v bds/cmd/maild
	GOPATH=$(PWD) go install -v bds/cmd/newmail
	GOPATH=$(PWD) go install -v bds/cmd/bdsconfig

clean:
	go clean -v
	rm -rf pkg

test:
	go test bds/lib/...
