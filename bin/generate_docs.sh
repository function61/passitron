#!/bin/bash -eu

export LC_ALL="C.UTF-8"
export LANG="C.UTF-8"

commandlineUserguide() {
	# because help text self reflects its binary name
	cp rel/pism_linux-amd64 pism

	cat << EOF > docs/user-guides/command-line.md
To receive help, just run:

\`\`\`
./pism --help
$(./pism --help)
\`\`\`

Any subcommand will also give you help:

\`\`\`
./pism server --help
$(./pism server --help)
\`\`\`

EOF

	# cleanup
	rm -f pism
}

commandlineUserguide


# internal dependencies

echo "\`\`\`" > docs/internal-dependencies.md
(cd cmd/pism && depth . | grep github.com/function61/pi-security-module/ | grep -v vendor) >> docs/internal-dependencies.md
echo "\`\`\`" >> docs/internal-dependencies.md

# mkdocs

rm -rf docs_ready/
cd docs/
mkdocs build --clean -f mkdocs.yml -d ../docs_ready

cd ../docs_ready/ && tar -czf docs.tar.gz *
