
VERSION := $(shell git describe --always --long --dirty)
DATE := $(shell date)

build:
	go build -ldflags "-s -w -X 'main.version=${VERSION}' -X 'main.date=${DATE}'"

install:
	go install -ldflags "-s -w -X 'main.version=${VERSION}' -X 'main.date=${DATE}'"



