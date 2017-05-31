#!/bin/bash -eu

docker run --cap-add=SYS_ADMIN -it -v "$(pwd):/app" --rm -p 8081:80 pi-security-module bash
