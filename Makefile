.PHONY: deps run test

deps:
	@go mod download

build: deps
	@go build

test:
	@go test