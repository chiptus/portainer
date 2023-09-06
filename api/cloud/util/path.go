package util

import (
	"os"
	"slices"
	"strings"
)

const pathKey = "PATH"

func PrependPathEnvironment(p string) {
	pathEnv := os.Getenv(pathKey)

	// does p exist in the path? very thorough check
	paths := strings.Split(pathEnv, string(os.PathListSeparator))
	if !slices.Contains(paths, p) {
		// prepend to original path and update environment
		pathEnv = p + string(os.PathListSeparator) + pathEnv
		os.Setenv(pathKey, pathEnv)
	}
}
