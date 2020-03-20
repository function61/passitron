#!/bin/bash -eu

source /build-common.sh

COMPILE_IN_DIRECTORY="cmd/passitron"
BINARY_NAME="passitron"
GOFMT_TARGETS="cmd/ pkg/"

# clean slate, because generated files rarely pass formatting check
cleanupGeneratedFiles() {
	rm -rf \
		docs_ready/ \
		pkg/**/*.gen.go
}

buildInternalDependenciesDocs() {
	echo "\`\`\`" > docs/internal-dependencies.md
	(cd cmd/passitron && depth . | grep github.com/function61/passitron/ | grep -v vendor) >> docs/internal-dependencies.md
	echo "\`\`\`" >> docs/internal-dependencies.md
}

generateCommandlineUserguideDocs() {
	# because help text self reflects its binary name
	cp rel/passitron_linux-amd64 passitron

	cat << EOF > docs/user-guides/command-line.md
To receive help, just run:

\`\`\`
./passitron --help
$(./passitron --help)
\`\`\`

Any subcommand will also give you help:

\`\`\`
./passitron server --help
$(./passitron server --help)
\`\`\`

EOF

	# cleanup
	rm -f passitron
}

if [ ! -n "${FASTBUILD:-}" ]; then
	cleanupGeneratedFiles
fi

standardBuildProcess

buildInternalDependenciesDocs

generateCommandlineUserguideDocs
