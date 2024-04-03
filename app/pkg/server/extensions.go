package server

import (
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

// findExtensionsInDir returns a list of all the extensions
func findExtensionsInDir(extDir string) ([]string, error) {
	log := zapr.NewLogger(zap.L())
	if extDir == "" {
		return nil, errors.New("extensions dir is empty")
	}
	entries, err := os.ReadDir(extDir)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", extDir)
	}

	extLocations := make([]string, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Extensions should contain a package.json file
		pkgFile := filepath.Join(extDir, entry.Name(), "package.json")

		_, err := os.Stat(pkgFile)
		if err != nil && os.IsNotExist(err) {
			log.Info("dir does not contain a package.json file; skipping it as an extension", "dir", entry.Name())
			continue
		}
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to stat %s", pkgFile)
		}

		extPath := filepath.Join(extDir, entry.Name())
		log.Info("Found extension", "dir", extPath)
		extLocations = append(extLocations, extPath)
	}
	return extLocations, nil
}
