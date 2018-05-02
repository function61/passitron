#!/bin/bash -eu

logBuild () {
	target="$1"

	echo "# Building $target"
}

downloadDependencies() {
	echo "# Downloading dependencies"

	dep ensure
}

codeGeneration() {
	echo "# Code generation"

	go generate
}

staticAnalysis() {
	echo "# Static analysis"
	
	go vet ./...
}

buildPublicFiles() {
	echo "# Building public files"

	(cd frontend/ && npm install)

	bin/tsc.sh

	bin/tslint.sh
}

packagePublicFiles() {
	echo "# Building rel/public.tar.gz"

	tar -czf rel/public.tar.gz public/
}

buildBinaries() {
	logBuild "linux-arm"

	GOOS=linux GOARCH=arm go build -o rel/pism_linux-arm

	logBuild "linux-amd64"

	GOOS=linux GOARCH=amd64 go build -o rel/pism_linux-amd64
}

uploadArtefacts() {
	echo "# Publishing build artefacts"

	# the CLI breaks automation unless opt-out..
	export JFROG_CLI_OFFER_CONFIG=false

	jfrog-cli bt upload \
		"--user=joonas" \
		"--key=$BINTRAY_APIKEY" \
		--publish=true \
		'rel/*' \
		"function61/pi-security-module/main/$FRIENDLY_REV_ID" \
		"$FRIENDLY_REV_ID/"
}

rm -rf rel
mkdir rel

downloadDependencies

codeGeneration

staticAnalysis

buildPublicFiles

packagePublicFiles

buildBinaries

uploadArtefacts
