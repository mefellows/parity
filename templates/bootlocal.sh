#!/bin/sh

cd /var/lib/boot2docker/

if [ ! -f "./mirror" ]; then
  wget https://github.com/mefellows/mirror/releases/download/{{.Version}}/linux_amd64.zip
  unzip linux_amd64.zip
fi

sudo cp -f ./mirror /usr/bin/mirror
./mirror-daemon.sh start
