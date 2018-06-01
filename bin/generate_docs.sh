#!/bin/bash -eu

go generate ./...

export LC_ALL="C.UTF-8"
export LANG="C.UTF-8"

# internal dependencies

echo "\`\`\`" > docs/internal-dependencies.md
(cd cmd/pism && depth . | grep github.com/function61/pi-security-module/ | grep -v vendor) >> docs/internal-dependencies.md
echo "\`\`\`" >> docs/internal-dependencies.md

# mkdocs

rm -rf docs_ready/
cd docs/
mkdocs build --clean -f mkdocs.yml -d ../docs_ready

cd ../docs_ready/ && tar -czf docs.tar.gz *
