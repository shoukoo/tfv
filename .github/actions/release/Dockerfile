FROM golang:latest

LABEL name="Golang Action"
LABEL maintainer="Andy Hsieh and Alex Lance"
LABEL version="0.0.1"
LABEL repository="https://github.com/shoukoo/tf-verifier"

LABEL com.github.actions.name="Golang Action"
LABEL com.github.actions.description="Run Golang commands"
LABEL com.github.actions.icon="mic"
LABEL com.github.actions.color="yellow"

RUN apt-get install git jq

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
