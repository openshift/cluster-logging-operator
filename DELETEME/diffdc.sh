#!/bin/sh
# Remove "add" blocks from diff output, keep delete/change blocks.

diff -b "$@" | awk '/^[0-9]+[dc][0-9]+/{p=1} /^[0-9]+a[0-9]+/{p=0} { if (p == 1) print}'
