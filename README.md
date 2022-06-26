# gmc (Go mod create)

`gmc` (Go mod create) is a CLI that creates Go modules.

## Install

```sh
$ go install github.com/jbrudvik/gmc@latest
```

## Create a package

```
$ gmc example.com/foo/bar
Creating Go module "example.com/foo/bar"...
Success! Created Go module "example.com/foo/bar" in new directory: bar
```

## Show help

```
$ gmc -h

NAME:
   gmc - (Go mod create) creates Go modules

USAGE:
   gmc [global options] [module name]

VERSION:
   v0.0.1

DESCRIPTION:
   gmc [module name] creates a directory containing:
   - go.mod            Go module metadata
   - .gitignore        ignores your module's binary
   - main.go           your module's first code
   - .nova (Optional)  Nova editor configuration

   This directory can be immediately built/run/installed using the `go` CLI.

   More information: https://github.com/jbrudvik/gmc

GLOBAL OPTIONS:
   --nova, -n     include Nova configuration (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```
