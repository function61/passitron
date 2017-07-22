FROM golang:1.8.0

# $ docker build -t pi-security-module .
# $ docker run --rm -it -p 8080:80 -p 8096:8096 -v "$(pwd):/go/src/github.com/function61/pi-security-module" pi-security-module

CMD bash

WORKDIR /go/src/github.com/function61/pi-security-module

RUN mkdir -p /tmp/glide \
	&& curl --location --fail https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-linux-amd64.tar.gz \
	| tar -C /tmp/glide -xzf - \
	&& mv /tmp/glide/linux-amd64/glide /usr/local/bin/ \
	&& rm -rf /tmp/glide

