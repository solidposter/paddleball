
VERSION := $(shell git describe --always --long --dirty)
DATE := $(shell date)

build:
	go build -ldflags "-X 'main.version=${VERSION}' -X 'main.date=${DATE}'" -o ./bin/paddleball ./cmd/paddleball

install:
	go install -ldflags "-X 'main.version=${VERSION}' -X 'main.date=${DATE}'" -o ./bin/paddleball ./cmd/paddleball



