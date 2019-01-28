FROM fn61/buildkit-golang:20190125_1536_1b2a32b5

WORKDIR /go/src/github.com/function61/pi-security-module

CMD bin/build.sh
