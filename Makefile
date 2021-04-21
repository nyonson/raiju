
BINDIR		:= $(DESTDIR)/usr/bin
VERSION		:= $(shell git describe --tags)
LANHOST 	:= lightning@gemini

.PHONY: test
test:
	@go test -cover ./...

.PHONY: build
build:
	@go build -ldflags="-X main.version=${VERSION}"  -o build/raiju cmd/raiju/main.go

install: test build
	@echo "installing"
	@install -D build/raiju -m 755 -t $(BINDIR)

uninstall:
	@echo "uninstalling"
	@rm -f $(BINDIR)/raiju

deploy-lan: test build
	@scp build/raiju $(LANHOST):~