VERSION := $(shell git describe --tags --always)

build:
	go build \
		-ldflags "-X main.Version=$(VERSION)" \
		-o bin/guiio