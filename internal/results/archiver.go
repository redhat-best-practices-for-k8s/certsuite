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

// generateZipFileName creates a timestamped filename for the results archive.
//
// It returns a string formatted as "<YYYYMMDDHHMMSS>.tar.gz", where the
// timestamp is generated from the current time at call time. This name is used
// when writing the compressed results file to disk.
func generateZipFileName() string {
	return time.Now().Format(tarGzFileNamePrefixLayout) + "-" + tarGzFileNameSuffix
}

// getFileTarHeader returns a tar header for the given file path.
//
// It takes a string representing the path to a file, retrieves its
// FileInfo via os.Stat, and creates a tar.Header using tar.FileInfoHeader.
// The function sets the header name to the provided path. If any step fails,
// it returns an error describing the issue. The returned *tar.Header can be
// used when adding files to a tar archive.
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

// CompressResultsArtifacts creates a compressed archive of result files.
//
// It accepts an output directory and a slice of file paths to include in the
// archive. The function generates a zip file name, writes each specified file
// into the archive, and returns the full path to the created zip file along
// with any error encountered during the process. If successful, the returned
// string is the absolute path to the new archive and the error is nil.
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
