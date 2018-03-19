#!/bin/bash -eu

logBuild () {
	target="$1"

	echo "# Building $target"
}

rm -rf rel
mkdir rel

echo "# Downloading dependencies"

dep ensure

echo "# Code generation and static analysis"

go generate

go vet

echo "# Building rel/public.tar.gz"

tar -czf rel/public.tar.gz public/

logBuild "linux-arm"

GOOS=linux GOARCH=arm go build -o rel/pism_linux-arm

logBuild "linux-amd64"

GOOS=linux GOARCH=amd64 go build -o rel/pism_linux-amd64

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
