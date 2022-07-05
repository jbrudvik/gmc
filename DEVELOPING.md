# Developing

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

1. Ensure all steps in [Before committing](#before-committing) section pass
1. [Create a new release](https://github.com/jbrudvik/gmc/releases/new) with:
   - Version: Incremented in format: vX.Y.Z
   - Release title: gmc `<version-from-last-step>`
   - Release notes
