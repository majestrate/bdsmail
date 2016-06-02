#/usr/bin/env bash
set -e
export GOPATH=./go

go get -u -v github.com/majestrate/bdsmail/cmd/bdsconfig
go get -u -v github.com/majestrate/bdsmail/cmd/bdsmail
cp "$GOPATH"/bin/bdsmail .
cp "$GOPATH"/bin/bdsconfig .
