# duf

[![Latest Release](https://img.shields.io/github/release/muesli/duf.svg)](https://github.com/muesli/duf/releases)
[![Build Status](https://github.com/muesli/duf/workflows/build/badge.svg)](https://github.com/muesli/duf/actions)
[![Go ReportCard](http://goreportcard.com/badge/muesli/duf)](http://goreportcard.com/report/muesli/duf)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/muesli/duf)

Disk Usage/Free Utility (Linux, BSD, macOS & Windows)

![duf](/duf.png)

## Features

- [x] User-friendly, colorful output
- [x] Adjusts to your terminal's width
- [x] Sort the results according to your needs
- [x] Groups & filters devices
- [x] Can conveniently output JSON

## Installation

### Packages

#### Linux
- Arch Linux: [duf](https://aur.archlinux.org/packages/duf/)
- Nix: `nix-env -iA nixpkgs.duf`
- [Packages](https://github.com/muesli/duf/releases) in Alpine, Debian & RPM formats

#### BSD
- FreeBSD: `pkg install duf`

#### macOS
- with [Homebrew](https://brew.sh/): `brew install duf`
- with [MacPorts](https://www.macports.org): `sudo port selfupdate && sudo port install duf`

#### Windows
- with [scoop](https://scoop.sh/): `scoop install duf`

#### Android
- Android (via termux): `pkg install duf`

### Binaries
- [Binaries](https://github.com/muesli/duf/releases) for Linux, FreeBSD, OpenBSD, macOS, Windows

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

If you supply arguments, duf will only list specific devices & mount points:

    duf /home /some/file

If you want to list everything (including pseudo, duplicate, inaccessible file systems):

    duf --all

You can show and hide specific tables:

    duf --only local,network,fuse,special,loops,binds
    duf --hide local,network,fuse,special,loops,binds

You can also show and hide specific filesystems:

    duf --only-fs tmpfs,vfat
    duf --hide-fs tmpfs,vfat

Sort the output:

    duf --sort size

Valid keys are: `mountpoint`, `size`, `used`, `avail`, `usage`, `inodes`,
`inodes_used`, `inodes_avail`, `inodes_usage`, `type`, `filesystem`.

Show or hide specific columns:

    duf --output mountpoint,size,usage

Valid keys are: `mountpoint`, `size`, `used`, `avail`, `usage`, `inodes`,
`inodes_used`, `inodes_avail`, `inodes_usage`, `type`, `filesystem`.

List inode information instead of block usage:

    duf --inodes

If duf doesn't detect your terminal's colors correctly, you can set a theme:

    duf --theme light

If you prefer your output as JSON:

    duf --json

## Troubleshooting

Users of `oh-my-zsh` should be aware that it already defines an alias called
`duf`, which you will have to remove in order to use `duf`:

    unalias duf
