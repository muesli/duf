# duf

[![Latest Release](https://img.shields.io/github/release/muesli/duf.svg)](https://github.com/muesli/duf/releases)
[![Build Status](https://github.com/muesli/duf/workflows/build/badge.svg)](https://github.com/muesli/duf/actions)
[![Go ReportCard](http://goreportcard.com/badge/muesli/duf)](http://goreportcard.com/report/muesli/duf)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/muesli/duf)

Disk Usage/Free Utility (currently Linux & macOS-only, support for BSDs soon)

![duf](/duf.png)

## Features

- [x] User-friendly, colorful output
- [x] Adjusts to your terminal's width
- [x] Sort the results according to your needs
- [x] Groups & filters devices
- [x] Can conveniently output JSON

## Installation

### Packages & Binaries

- Arch Linux: [duf](https://aur.archlinux.org/packages/duf/)
- [Binaries](https://github.com/muesli/duf/releases) for Linux

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

If you want to see all devices:

    duf --all

You can hide individual tables:

    duf --hide-local --hide-network --hide-fuse --hide-special --hide-loopback --hide-binds

Sort the output:

    # valid sort-keys are: mountpoint, size, used, avail, usage, type, filesystem
    duf --sort size

If you prefer your output as JSON:

    duf --json
