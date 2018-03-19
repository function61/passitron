#!/bin/bash -eu

# "3811ac3f4de838f51f877d55d74881ccb431d4b0" => "3811ac3f4de838f5"
CI_REVISION_ID_SHORT=${CI_REVISION_ID:0:16}

# "20171217_1632_3811ac3f4de838f5"
FRIENDLY_REV_ID="$(date +%Y%m%d_%H%M)_$CI_REVISION_ID_SHORT"

docker_run_contextless_build () {
	local dockerfile="$1"
	local image_name="$2"

	# Docker *requires* a directory that is used for build context. with the
	# build image we don't need any, so we'll create an empty directory for it.
	# IIRC it needs to be a subdirectory of our current working directory.
	mkdir -p empty_dir

	cp "$dockerfile" "empty_dir/${dockerfile}"

	docker build -f "empty_dir/${dockerfile}" -t "$image_name" empty_dir/
}

docker_run_contextless_build "Dockerfile.build" "pism-builder"

docker run \
	--rm \
	-it \
	--workdir "/go/src/github.com/function61/pi-security-module" \
	-v "$(pwd):/go/src/github.com/function61/pi-security-module" \
	-e "CI_REVISION_ID=$CI_REVISION_ID" \
	-e "FRIENDLY_REV_ID=$FRIENDLY_REV_ID" \
	-e "BINTRAY_APIKEY=$BINTRAY_APIKEY" \
	pism-builder \
	bin/build.sh
