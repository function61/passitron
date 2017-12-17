#!/bin/bash -eu

# "3811ac3f4de838f51f877d55d74881ccb431d4b0" => "3811ac3f4de838f5"
CI_REVISION_ID_SHORT=${CI_REVISION_ID:0:16}

# "20171217_1632_3811ac3f4de838f5"
FRIENDLY_REV_ID="$(date +%Y%m%d_%H%M)_$CI_REVISION_ID_SHORT"

# use (almost) empty directory as a build context for the build image
mkdir -p empty_dir

cp Dockerfile.build empty_dir/Dockerfile.build

docker build -f empty_dir/Dockerfile.build -t bob empty_dir/

docker run \
	--rm \
	-it \
	-v "$(pwd):/go/src/github.com/function61/pi-security-module" \
	-e "CI_REVISION_ID=$CI_REVISION_ID" \
	-e "FRIENDLY_REV_ID=$FRIENDLY_REV_ID" \
	-e "BINTRAY_APIKEY=$BINTRAY_APIKEY" \
	bob
