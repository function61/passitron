FROM fn61/buildkit-golang:20181005_1740_183e9622c00c5c6b

WORKDIR /go/src/github.com/function61/pi-security-module

CMD bin/build.sh

RUN curl -sL https://deb.nodesource.com/setup_8.x | bash - \
	&& apt-get install -y nodejs mkdocs
