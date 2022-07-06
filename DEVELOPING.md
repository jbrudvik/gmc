# Developing

## Setup

- Install Go 1.18: https://go.dev/doc/install
- Install [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports): `$ go install golang.org/x/tools/cmd/goimports@latest`

## Before committing

```sh
# Format all code correctly
$ goimports -w .

# Ensure code builds succesfully
$ go build

# Ensure code is free of common mistakes
$ go vet ./...

# Ensure all tests pass
$ go test ./...
```

## Releasing

1. Ensure build is passing: [![Build](https://github.com/jbrudvik/gmc/actions/workflows/build.yml/badge.svg)](https://github.com/jbrudvik/gmc/actions/workflows/build.yml)
1. [Create a new release](https://github.com/jbrudvik/gmc/releases/new) with:
   - Version: Incremented in format: vX.Y.Z
   - Release title: gmc `<version-from-last-step>`
   - Release notes
