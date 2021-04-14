
BINDIR         = $(DESTDIR)/usr/bin

.PHONY: test
test:
	@go test -cover ./...

.PHONY: build
build:
	@go build -o build/raiju cmd/raiju/main.go

install: test build
	@echo "installing"
	@install -D build/raiju -m 755 -t $(BINDIR)

uninstall:
	@echo "uninstalling"
	@rm -f $(BINDIR)/raiju