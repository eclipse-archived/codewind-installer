#!/bin/bash

cd cmd/cli
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cwctl-linux 
cd ../..

docker build -t cwctl-insecure .

rm cmd/cli/cwctl-linux

docker run -it -v /var/run/docker.sock:/var/run/docker.sock --rm cwctl-insecure ./cwctl --json start
