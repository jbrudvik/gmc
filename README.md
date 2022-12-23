# gmc (Go mod create)

`gmc` (Go mod create) is a CLI that creates Go modules so you can start coding ASAP

[![Build](https://github.com/jbrudvik/gmc/actions/workflows/build.yml/badge.svg)](https://github.com/jbrudvik/gmc/actions/workflows/build.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/jbrudvik/gmc.svg)](https://pkg.go.dev/github.com/jbrudvik/gmc)

## Usage

### Create a module as Git repository

```
$ gmc -g github.com/jbrudvik/mymodule
Creating Go module: github.com/jbrudvik/mymodule
- Created directory: mymodule
- Initialized Go module
- Created file     : mymodule/main.go
- Created file     : mymodule/.gitignore
- Initialized Git repository
- Created file     : mymodule/README.md
- Committed all files to Git repository
- Added remote for Git repository: git@github.com:jbrudvik/mymodule.git

Finished creating Go module: github.com/jbrudvik/mymodule

Next steps:
- Change into your module's directory: $ cd mymodule
- Run your module: $ go run .
- Create remote Git repository git@github.com:jbrudvik/mymodule.git: https://github.com/new
- Push to remote Git repository: $ git push -u origin main
- Start coding: $ vim .
```

### Show help

```
$ gmc -h
NAME:
   gmc - (Go mod create) creates Go modules so you can start coding ASAP

USAGE:
   gmc [global options] [module name]

VERSION:
   vX.Y.Z

DESCRIPTION:
   `gmc [module name]` creates a directory containing:
   - Go module metadata: go.mod
   - A place to start writing code: main.go
   - A .gitignore file

   This module can be immediately run:

       $ go run .
       hello, world!

   Optionally, the directory can also include:
   - Git repository setup with .gitignore, README.md

   More information: https://github.com/jbrudvik/gmc

GLOBAL OPTIONS:
   --git, -g      create as Git repository (default: false)
   --quiet, -q    silence output (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Install

### Required dependencies

- [Go 1.18](https://go.dev/doc/install)
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports): `$ go install golang.org/x/tools/cmd/goimports@latest`
- [Git](https://git-scm.com) (for Git-related features)

### Install gmc

```sh
$ go install github.com/jbrudvik/gmc@latest
```
