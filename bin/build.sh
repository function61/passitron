#!/bin/bash -eu

source /build-common.sh

COMPILE_IN_DIRECTORY="cmd/pism"
BINARY_NAME="pism"
GOFMT_TARGETS="cmd/ pkg/"

# clean slate, because generated files rarely pass formatting check
cleanupGeneratedFiles() {
	rm -rf \
		docs_ready/ \
		pkg/**/*.gen.go
}

buildInternalDependenciesDocs() {
	echo "\`\`\`" > docs/internal-dependencies.md
	(cd cmd/pism && depth . | grep github.com/function61/pi-security-module/ | grep -v vendor) >> docs/internal-dependencies.md
	echo "\`\`\`" >> docs/internal-dependencies.md
}

generateCommandlineUserguideDocs() {
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

if [ ! -n "${FASTBUILD:-}" ]; then
	cleanupGeneratedFiles
fi

standardBuildProcess

buildInternalDependenciesDocs

generateCommandlineUserguideDocs
