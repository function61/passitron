FROM node:7

# docker run --rm -it -v "$(pwd):/buildresult" node:7 bash

RUN git clone https://github.com/edi9999/jsqrcode.git /work \
	&& cd /work \
	&& npm install \
	&& npm run build \
	&& cp dist/index.js dist/index.js.map /buildresult
