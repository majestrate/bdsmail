GOPATH=$(PWD)

build:
	go install -v bds/cmd/maild
	go install -v bds/cmd/bdsconfig

clean:
	go clean -v
	rm -rf pkg
