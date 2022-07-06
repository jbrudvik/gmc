# gmc (Go mod create)

`gmc` (Go mod create) is a CLI that creates Go modules so you can start coding ASAP

## Usage

### Create a module as Git repository

```
$ gmc -g github.com/jbrudvik/foo
Creating Go module: github.com/jbrudvik/foo
- Created directory: foo
- Initialized Go module
- Created file     : foo/main.go
- Initialized Git repository
- Created file     : foo/.gitignore
- Created file     : foo/README.md
- Committed all files to Git repository
- Added remote for Git repository: git@github.com:jbrudvik/foo.git

Finished creating Go module: github.com/jbrudvik/foo

Next steps:
- Create remote Git repository git@github.com:jbrudvik/foo.git: https://github.com/new
- Push to remote Git repository: $ git push -u origin main
- Start coding: $ vim foo
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

   This module can be immediately run:

       $ go run .
       hello, world!

   Optionally, the directory can also include:
   - Git repository setup with .gitignore, README.md
   - Nova editor configuration to build/test/run natively

   More information: https://github.com/jbrudvik/gmc

GLOBAL OPTIONS:
   --git, -g      create as Git repository (default: false)
   --nova, -n     include Nova configuration (default: false)
   --quiet, -q    silence output (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Install

### Required dependencies

- [Go 1.18](https://go.dev/doc/install)
- [Git](https://git-scm.com) (for Git-related features)

### Install gmc

```sh
$ go install github.com/jbrudvik/gmc@latest
```
