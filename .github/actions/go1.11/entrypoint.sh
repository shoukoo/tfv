#!/bin/sh

set -exuo pipefail

export GO111MODULE=on
ls -al
make test
