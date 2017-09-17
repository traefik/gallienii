.PHONY: all

default: build

dependencies:
	dep ensure

build:
	go build