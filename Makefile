VERSION := $(shell git describe --tags --always)

.PHONY: all
all: build

.PHONY: build
build:
	@echo "===> GO Mod Check"
	@echo "===> Copy Env"
	cp -fp ./env/.env.json ./bin/

	@echo "===> Go Build"
	go build \
		-ldflags "-X main.Version=$(VERSION)" \
		-o bin/guiio ./backend/cmd/api
