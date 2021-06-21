#!/bin/sh
# set -e
# curl -o ./wait-for https://raw.githubusercontent.com/eficode/wait-for/master/wait-for
# chmod u+x ./wait-for
# ./wait-for database:5432 -t 15
go mod init
go mod tidy
go test -v . -run "'$1'"
# go test -v .
