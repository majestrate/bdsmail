#/usr/bin/env bash
set -e
export GOPATH=$PWD/go

go get -v github.com/majestrate/bdsmail/cmd/bdsconfig
go get -v github.com/majestrate/bdsmail/cmd/bdsmail
cp $GOPATH/go/bin/bdsmail $PWD
cp $GOPATH/go/bin/bdsconfig $PWD
