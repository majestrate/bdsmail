#/usr/bin/env bash
set -e

go get -u -v github.com/majestrate/bdsmail/cmd/bdsconfig
go get -u -v github.com/majestrate/bdsmail/cmd/bdsmail
cp go/bin/bdsmail .
cp go/bin/bdsconfig .
