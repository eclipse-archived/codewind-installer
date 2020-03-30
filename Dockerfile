FROM ubuntu

WORKDIR /app

RUN apt-get -y update 

RUN apt-get -y install docker.io

ADD ./cmd/cli/cwctl-linux /app/cwctl
