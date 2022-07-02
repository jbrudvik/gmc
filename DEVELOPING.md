# Developing

## Before committing

```sh
$ go fmt
$ go test ./... # All tests must pass
```

## Releasing

1. Ensure code is formatted (`gofmt -d .`) and tests are passing (`go test ./...`)
1. [Create a new release](https://github.com/jbrudvik/gmc/releases/new) with:
   - Version: Incremented in format: vX.Y.Z
   - Release title: gmc `<version-from-last-step>`
   - Release notes
