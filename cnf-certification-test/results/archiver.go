package results

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	tarGzFileNameSuffix = "cnf-test-results.tar.gz"
)

func generateZipFileName() string {
	dateTimePrefixFormat := "%02d%02d%02d-%02d%02d%02d"

	now := time.Now().UTC()

	return fmt.Sprintf(dateTimePrefixFormat+"-"+tarGzFileNameSuffix,
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())
}

// Helper function to get the tar file header from a file.
func getFileTarHeader(file string) (*tar.Header, error) {
	info, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info from %s: %v", file, err)
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to get file info header for %s: %v", file, err)
	}

	return header, nil
}

// Creates a zip file in the outputDir containing each file in the filePaths slice.
func CompressResultsArtifacts(outputDir string, filePaths []string) error {
	zipFileName := generateZipFileName()
	zipFilePath := filepath.Join(outputDir, zipFileName)

	logrus.Infof("Compressing results artifacts into %s", zipFilePath)
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return fmt.Errorf("failed creating tar.gz file %s in dir %s (filepath=%s): %v",
			zipFileName, outputDir, zipFilePath, err)
	}

	zipWriter := gzip.NewWriter(zipFile)
	defer zipWriter.Close()

	tarWriter := tar.NewWriter(zipWriter)
	defer tarWriter.Close()

	for _, file := range filePaths {
		logrus.Debugf("Zipping file %s", file)

		tarHeader, err := getFileTarHeader(file)
		if err != nil {
			return err
		}

		err = tarWriter.WriteHeader(tarHeader)
		if err != nil {
			return fmt.Errorf("failed to write tar header for %s: %v", file, err)
		}

		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", file, err)
		}

		if _, err = io.Copy(tarWriter, f); err != nil {
			return fmt.Errorf("failed to tar file %s: %v", file, err)
		}

		f.Close()
	}

	return nil
}
