VERSION := $(shell git describe --always --long --dirty)

OUTPUT_DIR := build
OUTPUT_NAME := paddleball

# Check the OS and set the output name
ifeq ($(OS), Windows_NT)
    OUTPUT_NAME := $(OUTPUT_DIR)/paddleball.exe
else
    OUTPUT_NAME := $(OUTPUT_DIR)/paddleball
endif


.DEFAULT_GOAL := build
.PHONY: fmt vet build

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	CGO_ENABLED=0 go build -ldflags "-s -w -X 'main.version=${VERSION}'" -o $(OUTPUT_NAME)


