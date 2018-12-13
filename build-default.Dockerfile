FROM fn61/buildkit-golang:20181204_1302_5eedb86addc826e7

WORKDIR /go/src/github.com/function61/pi-security-module

CMD bin/build.sh

RUN curl -sL https://deb.nodesource.com/setup_8.x | bash - \
	&& apt-get install -y nodejs mkdocs
