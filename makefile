SHELL := /bin/bash
BASEDIR = $(shell pwd)




.PHONY: allcheck
allcheck: govet gotest building


.PHONY: building 
building:
	go build -v -o fake_mqtt_device . 
.PHONY: govet
govet: 
	go vet  ./...
.PHONY: gotest
gotest: 
	go test -v ./...