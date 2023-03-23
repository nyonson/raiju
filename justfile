# default help menu lists tasks
@help:
  just --list --justfile {{justfile()}} --list-heading $'raiju\n'

# generate test boilerplate code including marked interface stubs and test tables for exported functions
@generate:
  go install github.com/cweill/gotests/gotests@latest
  go install github.com/matryer/moq@v0.3.1
  find . -type f -name "*.go" ! -name "*test.go" -exec gotests -exported -w '{}' \;  
  go generate ./...

# install the executable
@install:
  go install cmd/raiju/raiju.go

# publish the current commit with a tag
@publish tag message:
  git tag -a {{tag}} -m "{{message}}"
  git push origin {{tag}}
  podman build -t ghcr.io/nyonson/raiju:{{tag}} -f Containerfile .
  podman push ghcr.io/nyonson/raiju:{{tag}}

# test all the codes
@test:
  go test -cover ./...
