package tarextract

// hat tip https://gist.github.com/indraniel/1a91458984179ab4cf80

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

func ExtractTarGz(gzipStream io.Reader) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("gzip.NewReader() failed: %s", err.Error())
	}

	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("Next() failed: %s", err.Error())
		}

		if err := pathLooksDangerous(header.Name); err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				return fmt.Errorf("Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(header.Name)
			if err != nil {
				return fmt.Errorf("Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("Copy() failed: %s", err.Error())
			}
			outFile.Close() // defer would leak in a loop
		default:
			return fmt.Errorf("unknown type: %x in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

func pathLooksDangerous(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("pathLooksDangerous: %s", path)
	}

	return nil
}
