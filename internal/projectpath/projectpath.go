package projectpath

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	//nolint:gocritic
	Root = filepath.Join(filepath.Dir(b), "../..")
)
