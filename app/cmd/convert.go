package cmd

import (
	"fmt"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/application"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// NewConvertCmd create a convert command
func NewConvertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert <dir> ...",
		Short: "Convert all .foyle files in the specified directory into markdown files",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				log := zapr.NewLogger(zap.L())
				if len(args) == 0 {
					log.Info("convert takes at least one argument which should be the directory to search for foyle files.")
				}
				logVersion()

				app := application.NewApp()
				if err := app.LoadConfig(cmd); err != nil {
					return err
				}
				if err := app.SetupLogging(false); err != nil {
					return err
				}

				for _, dir := range args {
					if err := convertDir(dir); err != nil {
						return err
					}
				}

				return nil
			}()
			if err != nil {
				fmt.Printf("Error running convert;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

func convertDir(dir string) error {
	log := zapr.NewLogger(zap.L())

	dir, err := filepath.Abs(dir)
	if err != nil {
		return errors.Wrapf(err, "Failed to get absolute path for %s", dir)
	}

	if _, err := os.Stat(dir); err != nil {
		return errors.Wrapf(err, "Directory %s doesn't exist", dir)
	}

	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".foyle" {
			return nil
		}

		newPath := strings.TrimSuffix(path, ".foyle") + ".md"

		if _, err := os.Stat(newPath); err == nil {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "Failed to read file %s", path)
		}
		doc := &v1alpha1.Doc{}
		if err := protojson.Unmarshal(b, doc); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal doc %s", newPath)
		}
		// Convert the file
		mdText := docs.DocToMarkdown(doc)

		if err := os.WriteFile(newPath, []byte(mdText), 0777); err != nil {
			return errors.Wrapf(err, "Failed to write file %s", newPath)
		}
		log.Info("Converted file", "markdown", newPath)
		return nil
	})
}
