package tarextract

// hat tip https://gist.github.com/indraniel/1a91458984179ab4cf80

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
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
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("Copy() failed: %s", err.Error())
			}
		default:
			return fmt.Errorf("unknown type: %x in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}
