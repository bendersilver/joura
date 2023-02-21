#!/usr/bin/env bash

V="v0.0.1"

go build -ldflags "-s -w"  -o  ../pkg/joura-$V-$(lsb_release -sc)-$(uname -p)