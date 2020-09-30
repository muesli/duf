# duf

[![Latest Release](https://img.shields.io/github/release/muesli/duf.svg)](https://github.com/muesli/duf/releases)
[![Build Status](https://github.com/muesli/duf/workflows/build/badge.svg)](https://github.com/muesli/duf/actions)
[![Go ReportCard](http://goreportcard.com/badge/muesli/duf)](http://goreportcard.com/report/muesli/duf)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/muesli/duf)

Disk Usage/Free Utility (Linux, BSD & macOS)

![duf](/duf.png)

## Features

- [x] User-friendly, colorful output
- [x] Adjusts to your terminal's width
- [x] Sort the results according to your needs
- [x] Groups & filters devices
- [x] Can conveniently output JSON

## Installation

### Packages

- Arch Linux: [duf](https://aur.archlinux.org/packages/duf/)
- macOS:
  - with [Homebrew](https://brew.sh/): `brew install muesli/tap/duf`
  - with [MacPorts](https://www.macports.org): `sudo port selfupdate && sudo port install duf`
- Nix: `nix-env -iA nixpkgs.duf`
- [Packages](https://github.com/muesli/duf/releases) in Debian & RPM formats
- Android (via termux): `pkg install duf`

### Binaries

- [Binaries](https://github.com/muesli/duf/releases) for Linux, FreeBSD, macOS

### From source

Make sure you have a working Go environment (Go 1.12 or higher is required).
See the [install instructions](http://golang.org/doc/install.html).

Compiling duf is easy, simply run:

    git clone https://github.com/muesli/duf.git
    cd duf
    go build

## Usage

You can simply start duf without any command-line arguments:

    duf

If you want to list everything (including pseudo, duplicate, inaccessible file systems):

    duf --all

You can hide individual tables:

    duf --hide-local --hide-network --hide-fuse --hide-special --hide-loops --hide-binds

List inode information instead of block usage:

    duf --inodes

Sort the output:

    duf --sort size

Valid keys are: `mountpoint`, `size`, `used`, `avail`, `usage`, `inodes`,
`inodes_used`, `inodes_avail`, `inodes_usage`, `type`, `filesystem`.

Show or hide specific columns:

    duf --output mountpoint,size,usage

Valid keys are: `mountpoint`, `size`, `used`, `avail`, `usage`, `inodes`,
`inodes_used`, `inodes_avail`, `inodes_usage`, `type`, `filesystem`.

If you prefer your output as JSON:

    duf --json

## Troubleshooting

Users of `oh-my-zsh` should be aware that it already defines an alias called
`duf`, which you will have to remove in order to use `duf`:

    unalias duf
