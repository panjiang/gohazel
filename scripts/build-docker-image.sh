#!/bin/bash -e

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/gohazel main.go

docker build -f Dockerfile -t panjiang/gohazel ./

rm dist/gohazel