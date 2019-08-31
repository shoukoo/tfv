#!/bin/bash

set -exuo pipefail

export GO111MODULE=on
go get honnef.co/go/tools/cmd/staticcheck 
staticcheck ./...
