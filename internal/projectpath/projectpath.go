package projectpath

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	_, b, _, _ = runtime.Caller(0)
	Root       = getCorrectPath()
)

func getCorrectPath() string {
	if os.Getenv("UNIT_TEST") != "" || strings.HasSuffix(os.Args[0], ".test") || strings.Contains(os.Args[0], "/T/") {
		// Root folder of this project used for unit testing
		//nolint:gocritic
		return filepath.Join(filepath.Dir(b), "../..")
	}

	// The real working directory of anything non unit test related
	path, _ := os.Getwd()
	return path
}
