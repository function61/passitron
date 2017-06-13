#!/bin/sh -eu

logBuild () {
	target="$1"

	echo "# Building $target"

	mkdir -p "rel/$target"
}

echo "# Building rel/public.tar.gz"

rm -f rel/public.tar.gz

tar -czf rel/public.tar.gz public/

logBuild "linux-arm"

GOOS=linux GOARCH=arm go build -o rel/linux-arm/pism

logBuild "linux-x86_64"

GOOS=linux GOARCH=amd64 go build -o rel/linux-x86_64/pism

logBuild "windows-x86_64"

GOOS=windows GOARCH=amd64 go build -o rel/windows-x86_64/pism.exe
