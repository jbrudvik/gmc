# Developing

## Before committing

```sh
# Code must be formatted correctly
$ goimports -w .

# Code must build successfully
$ go build

# Code must be free of common mistakes
$ go vet ./...

# All tests must pass
$ go test ./...
```

## Releasing

1. Ensure all steps in [Before committing](#before-committing) section pass
1. [Create a new release](https://github.com/jbrudvik/gmc/releases/new) with:
   - Version: Incremented in format: vX.Y.Z
   - Release title: gmc `<version-from-last-step>`
   - Release notes
