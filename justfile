# default help menu lists tasks
@help:
  just --list --justfile {{justfile()}} --list-heading $'raiju\n'

# install the executable
@install:
  go install cmd/raiju/raiju.go

# test all the codes
@test:
  go test -cover ./...
