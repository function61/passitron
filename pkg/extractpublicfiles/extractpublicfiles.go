package extractpublicfiles

import (
	"context"
	"errors"
	"fmt"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/gokit/fileexists"
	"github.com/function61/pi-security-module/pkg/tarextract"
	"io"
	"log"
	"net/url"
	"os"
)

const (
	PublicFilesArchiveFilename = "public.tar.gz"
	publicFilesDirectory       = "public"
)

var errDownloadWithDevVersion = errors.New("public files dir not exists and not using released version - don't know how to fix this")

func BintrayDownloadUrl(user string, repo string, filePath string) string {
	return fmt.Sprintf(
		"https://bintray.com/%s/%s/download_file?file_path=%s",
		user,
		repo,
		url.QueryEscape(filePath))
}

func downloadPublicFiles(downloadUrl string, destination string, logger *log.Logger) error {
	if downloadUrl == "" {
		return errDownloadWithDevVersion
	}

	logger.Printf(
		"downloadPublicFiles: %s missing; downloading from %s",
		destination,
		downloadUrl)

	tempFilename := destination + ".dltemp"

	tempFile, err := os.Create(tempFilename)
	defer tempFile.Close()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), ezhttp.DefaultTimeout10s)
	defer cancel()

	resp, errHttp := ezhttp.Get(ctx, downloadUrl)
	if errHttp != nil {
		return errHttp
	}
	defer resp.Body.Close()

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		return err
	}

	tempFile.Close() // double close is intentional

	if err := os.Rename(tempFilename, destination); err != nil {
		return err
	}

	logger.Printf("downloadPublicFiles: %s succesfully downloaded", destination)

	return nil
}

func Run(downloadUrl string, archiveFilename string, logger *log.Logger) error {
	dirExists, err := fileexists.Exists(publicFilesDirectory)
	if err != nil {
		return err
	}
	if dirExists { // our job here is done
		return nil
	}

	archiveExists, err := fileexists.Exists(archiveFilename)
	if err != nil {
		return err
	}
	if !archiveExists {
		if err := downloadPublicFiles(downloadUrl, archiveFilename, logger); err != nil {
			return err
		}
	}

	logger.Printf("extractPublicFiles: extracting public files from %s", archiveFilename)

	f, err := os.Open(archiveFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tarextract.ExtractTarGz(f); err != nil {
		return err
	}

	return nil
}
