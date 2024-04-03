// main is a simple script to prepare a directory containing the actual assets for the front end code.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-logr/zapr"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

var (
	// excludedFiles is a list of files that shouldn't be copied over
	// we can't exclude the "out" directory because it contains assets for extensions such as markdown-language-features
	excludedFiles = []string{"src", "node_modules", "yarn.lock"}

	// excludedExtensions are a list of extensions we don't include because they aren't useful right now
	excludedExtensions = []string{"bat", "clojure", "coffeescript", "cpp", "csharp", "fsharp",
		"groovy", "handlebars", "hlsl", "ini", "java", "julia", "lua", "npm", "objective-c", "perl", "php", "powershell",
		"pug", "r", "razor", "ruby", "rust", "scss", "shaderlab", "search-result", "sql", "swift", "xml", "vb"}
)

func SetupLogging() error {
	// Use a non-json configuration configuration
	c := zap.NewDevelopmentConfig()

	// Use the keys used by cloud logging
	// https://cloud.google.com/logging/docs/structured-logging
	c.EncoderConfig.LevelKey = "severity"
	c.EncoderConfig.TimeKey = "time"
	c.EncoderConfig.MessageKey = "message"

	lvl := "info"
	zapLvl := zap.NewAtomicLevel()

	if err := zapLvl.UnmarshalText([]byte(lvl)); err != nil {
		return errors.Wrapf(err, "Could not convert level %v to ZapLevel", lvl)
	}

	c.Level = zapLvl
	newLogger, err := c.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize zap logger; error %v", err))
	}

	zap.ReplaceGlobals(newLogger)

	return nil
}

type extenstionType string

const (
	browser          extenstionType = "browser"
	theme            extenstionType = "theme"
	language         extenstionType = "language"
	notebookRenderer extenstionType = "notebookRenderer"
	unknown          extenstionType = "unknown"
)

// getExtensionType returns the type of extension based on the pkg file
func getExtensionType(pkg map[string]interface{}) extenstionType {
	if _, ok := pkg["browser"]; ok {
		return browser
	}

	contributesRaw, ok := pkg["contributes"]

	if !ok {
		return unknown
	}

	contributes, ok := contributesRaw.(map[string]interface{})

	if !ok {
		return unknown
	}

	if _, ok := contributes["themes"]; ok {
		return theme

	}
	if _, ok := contributes["languages"]; ok {
		return language
	}
	if _, ok := contributes["notebookRenderer"]; ok {
		return notebookRenderer
	}

	return unknown
}

// maybeCreateNLS makes sure a package has a nls.json file; I think this is a language file.
// if its missing we end up with 404s
func maybeCreateNLS(targetDir string) {
	file := filepath.Join(targetDir, "package.nls.json")
	_, err := os.Stat(file)
	if err == nil {
		return
	}
	log := zapr.NewLogger(zap.L())
	if os.IsNotExist(err) {
		log.Info("Creating nls.json file", "file", file)
		os.WriteFile(file, []byte("{}"), 0644)
		return
	}
}

// run the script.
// vscode path where vscode is built.
// out is the directory where the assets will be copied to
func run(vscode string, out string) error {
	log := zapr.NewLogger(zap.L())
	if err := SetupLogging(); err != nil {
		return err
	}

	if _, err := os.Stat(out); err == nil {
		log.Info("Deleting output directory", "dir", out)
		if err := os.RemoveAll(out); err != nil {
			return errors.Wrapf(err, "Failed to remove output directory %s", out)
		}
	}

	// Copy vscode assets
	assets := []string{"out-vscode-reh-web-min", "resources"}
	for _, a := range assets {
		srcPath := filepath.Join(vscode, a)
		dest := filepath.Join(out, a)
		log.Info("Copying directory", "src", srcPath, "dest", dest)
		if err := copy.Copy(srcPath, dest); err != nil {
			return err
		}
	}

	if err := copyExtensions(vscode, out); err != nil {
		return err
	}
	return nil
}

func copyExtensions(vscode string, out string) error {
	log := zapr.NewLogger(zap.L())
	extDir := filepath.Join(vscode, "extensions")

	// extensions should be copied into the extensions directory
	extOut := filepath.Join(out, "extensions")

	if _, err := os.Stat(extDir); err != nil {
		return errors.Wrapf(err, "Failed to stat extensions directory %s", extDir)
	}

	skipped := map[string]bool{}
	for _, e := range excludedExtensions {
		skipped[e] = true
	}

	allowedTypes := map[string]bool{}
	for _, t := range []string{"browser", "language", "notebookRenderer", "theme"} {
		allowedTypes[t] = true
	}
	err := filepath.WalkDir(extDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		name := filepath.Base(path)
		if _, ok := skipped[name]; ok {
			log.Info("Skipping extension", "extension", name)
			// No need to recourse into this directory
			return filepath.SkipDir
		}
		pkgFile := filepath.Join(path, "package.json")
		if _, err := os.Stat(pkgFile); err != nil {
			if os.IsNotExist(err) {
				// N.B. we need to recourse into subdirectories because this branch will be hit for the "extensions"
				// directory and we need to recourse into the subdirectories which are the actual extensions.
				return nil
			}
			return err
		}

		pkgData, err := os.ReadFile(pkgFile)
		if err != nil {
			return err
		}
		var pkg map[string]interface{}
		if err := json.Unmarshal(pkgData, &pkg); err != nil {
			return err
		}
		extType := getExtensionType(pkg)
		if _, ok := allowedTypes[string(extType)]; !ok {
			log.Info("Skipping extension of unallowed type", "extension", name, "type", extType)
			return nil
		}

		targetDir := filepath.Join(extOut, name)
		if _, err := os.Stat(targetDir); err == nil {
			log.Info("Removing extension dir", "extension", name, "dir", targetDir)
			if err := os.RemoveAll(targetDir); err != nil {
				return err
			}
		}
		log.Info("Copying extension", "name", name, "src", path, "dest", targetDir)
		if err := copy.Copy(path, targetDir); err != nil {
			return err
		}

		maybeCreateNLS(targetDir)

		return filepath.SkipDir
	})

	return err
}

func main() {
	var vscode string
	var out string
	flag.StringVar(&vscode, "vscode", "", "path to the vscode directory")
	flag.StringVar(&out, "out", "", "path to story the assets in")
	flag.Parse()

	if err := run(vscode, out); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
