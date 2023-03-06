#!/usr/bin/env bash

VER='0.1.3'



DIR=dpkg/usr/bin
if [ "$(ls -A $DIR)" ]; then
    rm $DIR/*
fi

sed -ri "s/^Version: .*$/Version: $VER/g" dpkg/DEBIAN/control
sed -ri "s/^const Version = .*$/const Version = \"$VER\"/g" cmd/main.go

go build -ldflags "-s -w"  -o  $DIR/joura cmd/main.go

dpkg-deb --build dpkg/ joura_"$VER"-0$(lsb_release -sc)_$(uname -p).deb