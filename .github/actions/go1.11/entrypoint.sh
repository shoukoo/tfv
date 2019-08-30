#!/bin/sh

set -exuo pipeline

export GO111MODULE=on
make test
