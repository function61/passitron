#!/bin/sh -eu

logBuild () {
	target="$1"

	echo "# Building $target"
}

rm -rf rel/

mkdir rel

echo "# Building rel/public.tar.gz"

tar -czf rel/public.tar.gz public/

logBuild "linux-arm"

GOOS=linux GOARCH=arm go build -o rel/pism_linux-arm

logBuild "linux-amd64"

GOOS=linux GOARCH=amd64 go build -o rel/pism_linux-amd64

logBuild "windows-amd64"

GOOS=windows GOARCH=amd64 go build -o rel/pism_windows-amd64.exe
