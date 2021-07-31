# DESTDIR is a makefile convention for install/uninstall targets
BINDIR		:= $(DESTDIR)/usr/bin
VERSION		:= $(shell git describe --tags)

## test: run go tests
.PHONY: test
test:
	@go test -cover ./...

## build: build the executable
.PHONY: build
build:
	@go build -ldflags="-X main.version=${VERSION}"  -o build/raiju cmd/raiju/main.go

## install: install the executable into BINDIR
.PHONY: install
install: test build
	@echo "installing"
	@install -D build/raiju -m 755 -t $(BINDIR)

## uninstall: remove the installed executable
.PHONY: uninstall
uninstall:
	@echo "uninstalling"
	@rm -f $(BINDIR)/raiju

## deploy: deploy the executable to a HOST at ~
.PHONY: deploy
deploy: test build
	@test -n "$(DEPLOY_HOST)" || (echo 'A DEPLOY_HOST must be defined for deploy. Ex: make deploy DEPLOY_HOST=gemini' && exit 1)
	@scp build/raiju $(DEPLOY_HOST):~

## help: print help message
.DEFAULT_GOAL := help
.PHONY: help
help: Makefile
	@echo "MASH"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'