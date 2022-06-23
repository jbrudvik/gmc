# gmc (Go mod create)

`gmc` is a CLI that creates a Go module starting point.

## Usage

```
$ gmc

NAME:
   gmc - (Go mod create) Creates a Go module starting point

USAGE:
   gmc [global options] [module name]

VERSION:
   v0.0.1

DESCRIPTION:
   gmc creates a directory containing:
   - go.mod            Go module metadata
   - .gitignore        ignores your module's binary
   - main.go           a starting place for your module's code
   - .nova (Optional)  Nova editor configuration

   This new directory can be immediately built/run/installed using the go CLI.

   More information: https://github.com/jbrudvik/gmc

GLOBAL OPTIONS:
   --nova, -n     include Nova configuration (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Install

```sh
$ go install github.com/jbrudvik/gmc
```
