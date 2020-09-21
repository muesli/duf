# duf

[![Latest Release](https://img.shields.io/github/release/muesli/duf.svg)](https://github.com/muesli/duf/releases)
[![Build Status](https://github.com/muesli/duf/workflows/build/badge.svg)](https://github.com/muesli/duf/actions)
[![Go ReportCard](http://goreportcard.com/badge/muesli/duf)](http://goreportcard.com/report/muesli/duf)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/muesli/duf)

Disk Usage/Free Utility (currently Linux-only, support for BSDs soon)

![duf](/duf.png)

## Usage

You can simply start duf without any command-line arguments:

    duf

If you want to see all devices:

    duf -all

You can hide individual tables:

    duf -hide-local -hide-network -hide-fuse -hide-special
