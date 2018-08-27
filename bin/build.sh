#!/bin/bash -eu

run() {
	fn="$1"

	echo "# $fn"

	"$fn"
}

downloadDependencies() {
	dep ensure
}

codeGeneration() {
	go generate ./...
}

unitTests() {
	go test ./...
}

staticAnalysis() {
	go vet ./...
}

buildPublicFiles() {
	# --no-bin-links to work across filesystems, possibly on Win-Linux development with fileshares
	(cd frontend/ && npm install --no-bin-links)

	bin/tsc.sh

	bin/tslint.sh
}

packagePublicFiles() {
	tar -czf rel/public.tar.gz public/
}

buildLinuxArm() {
	(cd cmd/pism && GOOS=linux GOARCH=arm go build -o ../../rel/pism_linux-arm)
}

buildLinuxAmd64() {
	(cd cmd/pism && GOOS=linux GOARCH=amd64 go build -o ../../rel/pism_linux-amd64)
}

buildAndDeployDocs() {
	bin/generate_docs.sh

	mc config host add s3 https://s3.amazonaws.com "$AWS_ACCESS_KEY_ID" "$AWS_SECRET_ACCESS_KEY" S3v4

	mc cp --json --no-color docs_ready/docs.tar.gz s3/docs.function61.com/_packages/pi-security-module.tar.gz
}

uploadBuildArtefacts() {
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

run downloadDependencies

run codeGeneration

run staticAnalysis

run unitTests

run buildPublicFiles

run packagePublicFiles

run buildLinuxArm

run buildLinuxAmd64

run buildAndDeployDocs

run uploadBuildArtefacts
