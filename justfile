# default help menu lists tasks
@help:
  just --list --justfile {{justfile()}} --list-heading $'raiju\n'

# test all the codes
@test:
  go test -cover ./...
