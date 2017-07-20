FROM golang:1.8.0

# $ docker build -t pi-security-module .
# $ docker run --rm -it -p 8080:80 -p 8096:8096 -v "$(pwd):/app" pi-security-module bash

CMD mkdir -p /go/src/github.com/function61 && \
	ln -s /app /go/src/github.com/function61/pi-security-module

WORKDIR /app
