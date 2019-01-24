FROM fn61/buildkit-golang:20190108_1812_e64c80f1

WORKDIR /go/src/github.com/function61/pi-security-module

CMD bin/build.sh

RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - \
	&& echo "deb http://dl.yarnpkg.com/debian/ stable main" > /etc/apt/sources.list.d/yarn.list \
	&& curl -sL https://deb.nodesource.com/setup_8.x | bash - \
	&& apt install -y nodejs yarn
