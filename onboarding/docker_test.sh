#!/bin/sh

docker run -i -t -v "${GOPATH}/src/github.com/samsung-cnct/:/go/src/github.com/samsung-cnct/" docker.io/library/golang:1.8 \
  make -C src/github.com/samsung-cnct/technical-on-boarding/onboarding/ lint test
