# Developing

## Before committing

```sh
$ go fmt
$ go test ./... # All tests must pass
```

## Releasing

Releases are manually created with these steps:

1. Ensure code is formatted (`gofmt -d .`) and tests are passing (`go test`)
1. Increment the version in `cli/cli.go`
1. [Draft a new release](https://github.com/jbrudvik/note/releases/new) with a new tag that matches the version from the previous step
1. Add release notes since last release
1. Publish the release
