
VERSION := $(shell git describe --always --long --dirty)

build:
	go build -ldflags "-s -w -X 'main.version=${VERSION}'"

install:
	go install -ldflags "-s -w -X 'main.version=${VERSION}'"



