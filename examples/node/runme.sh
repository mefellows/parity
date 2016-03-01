#!/bin/bash

docker build -t parity-node .
docker run -it -v $PWD:/usr/src/app -p 3000:3000 parity-node bash
