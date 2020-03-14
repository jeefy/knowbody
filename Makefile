#!make
include .env
export $(shell sed 's/=.*//' .env)

.PHONY: check run

check: 
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.23.8
	golangci-lint run -c .golang-ci.yml ./... 
	go run cmd/main.go --lint

test:
	go test -v ./...

run:
	source .env
	go run cmd/main.go

build:
	go build -o out/knowbody cmd/main.go

build-image:
	docker build -t knowbody .