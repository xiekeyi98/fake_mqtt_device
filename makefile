SHELL := /bin/bash
BASEDIR = $(shell pwd)




.PHONY: allcheck
allcheck: govet gotest


.PHONY: govet
govet: 
	go vet ./...
.PHONY: gotest
gotest: 
	go test ./...