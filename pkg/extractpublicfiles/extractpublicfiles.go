package extractpublicfiles

import (
	"errors"
	"github.com/function61/pi-security-module/pkg/tarextract"
	"github.com/function61/pi-security-module/pkg/version"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	publicFilesArchive   = "public.tar.gz"
	publicFilesDirectory = "public"
)

var errDownloadWithDevVersion = errors.New("public files dir not exists and not using released version - don't know how to fix this")

func publicFilesDownloadUrl(versionStr string) string {
	return "https://bintray.com/function61/pi-security-module/download_file?file_path=" + versionStr + "%2Fpublic.tar.gz"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil { // exists
		return true
	}

	// we're only expecting "not exists" error - anything else is an actual error
	if !os.IsNotExist(err) {
		log.Fatalf("fileExists: unexpected error: %s", err.Error())
	}

	return false
}

func downloadPublicFiles() error {
	if version.IsDevVersion() {
		return errDownloadWithDevVersion
	}

	downloadUrl := publicFilesDownloadUrl(version.Version)

	log.Printf(
		"extractPublicFiles: %s missing; downloading from %s",
		publicFilesArchive,
		downloadUrl)

	tempFilename := publicFilesArchive + ".dltemp"

	tempFile, err := os.Create(tempFilename)
	defer tempFile.Close()
	if err != nil {
		return err
	}

	resp, errHttp := http.Get(downloadUrl)
	if errHttp != nil {
		return errHttp
	}
	defer resp.Body.Close()

	if _, errHttpDownload := io.Copy(tempFile, resp.Body); errHttpDownload != nil {
		return errHttpDownload
	}

	if errRename := os.Rename(tempFilename, publicFilesArchive); errRename != nil {
		return errRename
	}

	log.Printf("extractPublicFiles: %s succesfully downloaded", publicFilesArchive)

	return nil
}

func Run() error {
	// our job here is done
	if fileExists(publicFilesDirectory) {
		return nil
	}

	if !fileExists(publicFilesArchive) {
		if err := downloadPublicFiles(); err != nil {
			return err
		}
	}

	log.Printf("extractPublicFiles: extracting public files from %s", publicFilesArchive)

	f, err := os.Open(publicFilesArchive)
	if err != nil {
		log.Fatalf("extractPublicFiles: failed to open %s: %s", publicFilesArchive, err.Error())
	}
	defer f.Close()

	tarextract.ExtractTarGz(f)

	return nil
}
