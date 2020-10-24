#!/bin/bash -e

[ -z "$VERSION" ] && echo "Need to set VERSION" && exit 1;

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o gohazel main.go

zip -r gohazel-${VERSION}-linux-x86_64.zip gohazel config.yml

rm gohazel