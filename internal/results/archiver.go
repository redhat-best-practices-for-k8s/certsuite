package results

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

const (
	// tarGz file prefix layout format: YearMonthDay-HourMinSec
	tarGzFileNamePrefixLayout = "20060102-150405"
	tarGzFileNameSuffix       = "cnf-test-results.tar.gz"
)

// generateZipFileName creates a timestamped name for the archive file
//
// The function generates a string by formatting the current time with a
// predefined layout and appending a suffix to produce a unique filename. It
// uses the system clock to ensure each call returns a different value, suitable
// for naming compressed result artifacts. The returned string is later combined
// with a directory path to create the full file location.
func generateZipFileName() string {
	return time.Now().Format(tarGzFileNamePrefixLayout) + "-" + tarGzFileNameSuffix
}

// getFileTarHeader Creates a tar header for a given file
//
// The function retrieves the file’s metadata using the operating system, then
// converts that information into a tar header structure suitable for archiving.
// It returns the header or an error if either the stat call or the conversion
// fails.
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

// CompressResultsArtifacts Creates a compressed archive of specified files
//
// The function builds a zip file in the given output directory, including each
// path from the slice. It streams each file into a tar writer wrapped by gzip
// for compression, handling errors during header creation or file access. The
// absolute path to the resulting archive is returned.
func CompressResultsArtifacts(outputDir string, filePaths []string) (string, error) {
	zipFileName := generateZipFileName()
	zipFilePath := filepath.Join(outputDir, zipFileName)

	log.Info("Compressing results artifacts into %s", zipFilePath)
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return "", fmt.Errorf("failed creating tar.gz file %s in dir %s (filepath=%s): %v",
			zipFileName, outputDir, zipFilePath, err)
	}

	zipWriter := gzip.NewWriter(zipFile)
	defer zipWriter.Close()

	tarWriter := tar.NewWriter(zipWriter)
	defer tarWriter.Close()

	for _, file := range filePaths {
		log.Debug("Zipping file %s", file)

		tarHeader, err := getFileTarHeader(file)
		if err != nil {
			return "", err
		}

		err = tarWriter.WriteHeader(tarHeader)
		if err != nil {
			return "", fmt.Errorf("failed to write tar header for %s: %v", file, err)
		}

		f, err := os.Open(file)
		if err != nil {
			return "", fmt.Errorf("failed to open file %s: %v", file, err)
		}

		if _, err = io.Copy(tarWriter, f); err != nil {
			return "", fmt.Errorf("failed to tar file %s: %v", file, err)
		}

		f.Close()
	}

	// Create fully qualified path to the zip file
	zipFilePath, err = filepath.Abs(zipFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %v", zipFilePath, err)
	}

	// Return the entire path to the zip file
	return zipFilePath, nil
}
