#!/bin/bash -eu

buildFrontend() {
	source /build-common.sh

	standardBuildProcess "frontend"
}

packagePublicFiles() {
	tar -czf rel/public.tar.gz public/
}

(cd frontend/ && buildFrontend)

packagePublicFiles
