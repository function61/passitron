package main

import (
	"github.com/function61/pi-security-module/util/tarextract"
	"log"
	"os"
)

const (
	publicFilesArchive   = "public.tar.gz"
	publicFilesDirectory = "public"
)

func extractPublicFiles() {
	_, err := os.Stat(publicFilesDirectory)
	if err == nil { // dir exists
		return
	}
	if !os.IsNotExist(err) {
		log.Fatalf("extractPublicFiles: unexpected error: %s", err.Error())
	}

	log.Printf("extractPublicFiles: extracting public files from %s", publicFilesArchive)

	f, err := os.Open(publicFilesArchive)
	if err != nil {
		log.Fatalf("extractPublicFiles: failed to open %s: %s", publicFilesArchive, err.Error())
	}
	defer f.Close()

	tarextract.ExtractTarGz(f)
}
