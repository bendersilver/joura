#!/usr/bin/env bash

DIR="/usr/local/joura/v0.0.1"
BIN="$DIR"/joura

USR_BIN="/usr/bin/joura"


if [ $(dpkg-query -W -f='${Status}' libsystemd-dev 2>/dev/null | grep -c "ok installed") -eq 0 ]
then
        sudo apt install libsystemd-dev -y
fi

sudo mkdir -p "$DIR"
sudo mkdir -p "/etc/joura"


if systemctl is-active --quiet joura
then
        sudo service joura stop
fi


if [ ! -f "$BIN" ]
then
        sudo wget https://github.com/bendersilver/joura/releases/download/v0.0.1/joura-v0.0.1 -q --show-progress -O "$BIN"
fi

if [ ! -f "$DIR"/user.conf ]
then
        sudo wget https://github.com/bendersilver/joura/releases/download/v0.0.1/user.conf -q --show-progress -O "$DIR"/user.conf
fi

if [ ! -f /etc/joura/user.conf ]
then
        sudo cp "$DIR"/user.conf /etc/joura/user.conf
fi

if [ ! -f /etc/systemd/system/joura.service ]
then
        sudo rm "$DIR"/joura.service
        sudo wget https://github.com/bendersilver/joura/releases/download/v0.0.1/joura.service -q --show-progress -O "$DIR"/joura.service
        sudo cp "$DIR"/joura.service /etc/systemd/system/
        sudo systemctl enable joura.service
        sudo systemctl daemon-reload
fi

