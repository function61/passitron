#!/bin/bash -eu

source /build-common.sh

COMPILE_IN_DIRECTORY="cmd/pism"
BINARY_NAME="pism"
BINTRAY_PROJECT="function61/pi-security-module"
GOFMT_TARGETS="cmd/ pkg/"

# clean slate, because generated files rarely pass formatting check
cleanupGeneratedFiles() {
	rm -rf \
		docs_ready/ \
		pkg/apitypes/apitypes.go \
		pkg/apitypes/restendpoints.go \
		pkg/commandhandlers/commanddefinitions.go \
		pkg/domain/consts-and-enums.go \
		pkg/domain/events.go
}

buildPublicFiles() {
	# --no-bin-links to work across filesystems, possibly on Win-Linux development with fileshares
	(cd frontend/ && npm install --no-bin-links)

	# apparently --no-bin-links leaves executable bits off of these o_O
	chmod +x frontend/node_modules/typescript/bin/tsc frontend/node_modules/tslint/bin/tslint

	bin/tsc.sh

	bin/tslint.sh
}

packagePublicFiles() {
	tar -czf rel/public.tar.gz public/
}

buildAndDeployDocs() {
	bin/generate_docs.sh

	if [ "${PUBLISH_ARTEFACTS:-''}" = "true" ]; then
		mc config host add s3 https://s3.amazonaws.com "$AWS_ACCESS_KEY_ID" "$AWS_SECRET_ACCESS_KEY" S3v4

		mc cp --json --no-color docs_ready/docs.tar.gz s3/docs.function61.com/_packages/pi-security-module.tar.gz
	fi
}

hook_unitTests_after() {
	buildstep buildPublicFiles

	buildstep packagePublicFiles
}

cleanupGeneratedFiles

standardBuildProcess

buildAndDeployDocs
