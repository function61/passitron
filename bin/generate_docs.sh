#!/bin/bash -eu

go generate ./...

export LC_ALL="C.UTF-8"
export LANG="C.UTF-8"

rm -rf docs_ready/
cd docs/
mkdocs build --clean -f mkdocs.yml -d ../docs_ready

cd ../docs_ready/ && tar -czf docs.tar.gz *
