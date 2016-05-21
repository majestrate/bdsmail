#/usr/bin/env bash
set -e
export GOPATH=$PWD/go

go get -u -v github.com/majestrate/bdsmail/cmd/bdsconfig
go get -u -v github.com/majestrate/bdsmail/cmd/bdsmail
cp $GOPATH/bin/bdsmail $PWD
cp $GOPATH/bin/bdsconfig $PWD
