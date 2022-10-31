# duf

[![Latest Release](https://img.shields.io/github/release/muesli/duf.svg?style=for-the-badge)](https://github.com/muesli/duf/releases)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](https://pkg.go.dev/github.com/muesli/duf)
[![Software License](https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge)](/LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/muesli/duf/build?style=for-the-badge)](https://github.com/muesli/duf/actions)
[![Go ReportCard](https://goreportcard.com/badge/github.com/muesli/duf?style=for-the-badge)](https://goreportcard.com/report/muesli/duf)

Disk Usage/Free Utility (Linux, BSD, macOS & Windows)

![duf](/duf.png)

## Features

- [x] User-friendly, colorful output
- [x] Adjusts to your terminal's theme & width
- [x] Sort the results according to your needs
- [x] Groups & filters devices
- [x] Can conveniently output JSON

## Installation

### Packages

#### Linux
- Arch Linux: `pacman -S duf`
- Ubuntu 22.04 / Debian unstable: `apt install duf`
- Nix: `nix-env -iA nixpkgs.duf`
- Void Linux: `xbps-install -S duf`
- [Packages](https://github.com/muesli/duf/releases) in Alpine, Debian & RPM formats

#### BSD
- FreeBSD: `pkg install duf`
- OpenBSD: `pkg_add duf`

#### macOS
- with [Homebrew](https://brew.sh/): `brew install duf`
- with [MacPorts](https://www.macports.org): `sudo port selfupdate && sudo port install duf`

#### Windows
- with [Chocolatey](https://chocolatey.org/): `choco install duf`
- with [scoop](https://scoop.sh/): `scoop install duf`

#### Android
- Android (via termux): `pkg install duf`

### Binaries
- [Binaries](https://github.com/muesli/duf/releases) for Linux, FreeBSD, OpenBSD, macOS, Windows

### From source

Make sure you have a working Go environment (Go 1.16 or higher is required).
See the [install instructions](https://golang.org/doc/install.html).

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

### Filtering

You can show and hide specific tables:

    duf --only local,network,fuse,special,loops,binds
    duf --hide local,network,fuse,special,loops,binds

You can also show and hide specific filesystems:

    duf --only-fs tmpfs,vfat
    duf --hide-fs tmpfs,vfat

...or specific mount points:

    duf --only-mp /,/home,/dev
    duf --hide-mp /,/home,/dev

Wildcards inside quotes work:

    duf --only-mp '/sys/*,/dev/*'

### Display options

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

### Color-coding & Thresholds

duf highlights the availability & usage columns in red, green, or yellow,
depending on how much space is still available. You can set your own thresholds:

    duf --avail-threshold="10G,1G"
    duf --usage-threshold="0.5,0.9"

### Bonus

If you prefer your output as JSON:

    duf --json

## Troubleshooting

Users of `oh-my-zsh` should be aware that it already defines an alias called
`duf`, which you will have to remove in order to use `duf`:

    unalias duf

## Feedback

Got some feedback or suggestions? Please open an issue or drop me a note!

* [Twitter](https://twitter.com/mueslix)
* [The Fediverse](https://mastodon.social/@fribbledom)
