
BINDIR		:= $(DESTDIR)/usr/bin
VERSION		:= $(shell git describe --tags)
HOST	 	:= lightning@gemini

.PHONY: test
test:
	@go test -cover ./...

.PHONY: build
build:
	@go build -ldflags="-X main.version=${VERSION}"  -o build/raiju cmd/raiju/main.go

.PHONY: install
install: test build
	@echo "installing"
	@install -D build/raiju -m 755 -t $(BINDIR)

.PHONY: uninstall
uninstall:
	@echo "uninstalling"
	@rm -f $(BINDIR)/raiju

.PHONY: deploy
deploy: test build
	@scp build/raiju $(HOST):~