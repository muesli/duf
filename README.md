# duf

Disk Usage/Free Utility (currently Linux-only, support for BSDs soon)

![duf](/duf.png)

## Usage

You can simply start duf without any command-line arguments:

    duf

If you want to see all devices:

    duf -all

You can hide individual tables:

    duf -hide-local -hide-network -hide-fuse -hide-special
