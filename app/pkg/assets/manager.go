package assets

import (
	"archive/tar"
	"context"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-logr/zapr"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/hydros/pkg/files"
	"github.com/jlewi/hydros/pkg/images"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	// Constants representing subdirectories for different assets
	vscode    = "vscode"
	extension = "foyle"

	defaultVSCodeImage = "ghcr.io/jlewi/vscode-web-assets"
	defaultFoyleImage  = "ghcr.io/jlewi/foyle-vscode-ext"
)

// Manager is a struct that manages assets
type Manager struct {
	config      config.Config
	downloadDir string
}

// NewManager creates a new manager
func NewManager(config config.Config) (*Manager, error) {
	return &Manager{
		config: config,
	}, nil
}

type asset struct {
	source      string
	stripPrefix string
}

// Download downloads the assets
func (m *Manager) Download(ctx context.Context, tag string) error {
	log := logs.FromContext(ctx)

	// Map from the name of the asset to the source of the location
	assets := map[string]*asset{
		vscode: {
			source:      defaultVSCodeImage,
			stripPrefix: "assets",
		},
		extension: {
			source:      defaultFoyleImage,
			stripPrefix: "foyle",
		},
	}

	// If any assets are specified in the config file then the should override the defaults
	if m.config.Assets != nil {
		if m.config.Assets.VSCode != nil && m.config.Assets.VSCode.URI != "" {
			assets[vscode].source = m.config.Assets.VSCode.URI
		}
		if m.config.Assets.FoyleExtension != nil && m.config.Assets.FoyleExtension.URI != "" {
			assets[extension].source = m.config.Assets.FoyleExtension.URI
		}
	}

	if m.downloadDir == "" {
		tDir, err := os.MkdirTemp("", "temporaryAssets")
		if err != nil {
			return err
		}
		m.downloadDir = tDir
	}
	log.Info("Downloading assets", "downloadDir", m.downloadDir)

	if err := os.MkdirAll(m.downloadDir, 0755); err != nil {
		return errors.Wrapf(err, "Error creating directory %v", m.downloadDir)
	}
	for name, a := range assets {
		assetDir := filepath.Join(m.config.GetAssetsDir(), name)

		source := a.source
		u, err := url.Parse(source)
		if err != nil {
			return errors.Wrapf(err, "Error parsing source %v", source)
		}

		switch u.Scheme {
		case "":
		case "docker":
		default:
			return errors.Errorf("URI %v has unsupported scheme %v is not supported. Currently the only supported scheme is docker:// for images", source, u.Scheme)

		}

		uri := u.Path

		if uri == "" {
			return errors.Errorf("Asset %s has an empty source", name)
		}

		uri, err = resolveTag(uri, tag)
		if err != nil {
			return errors.Wrapf(err, "Error resolving tag for asset %v", name)
		}

		// TODO(jeremy): Should we check if its an empty directory
		if _, err := os.Stat(assetDir); err == nil {
			log.Info("Asset already exists", "assetDir", assetDir, "name", name, "source", source)
			continue
		}

		tarBall := filepath.Join(m.downloadDir, name+".tar.gz")

		if _, err := os.Stat(tarBall); err == nil {
			log.Info("Tarball already exists", "tarBall", tarBall)
			continue
		}
		log.Info("Downloading asset", "name", name, "source", source, "tarBall", tarBall)
		// Download the asset

		if err := images.ExportImage(uri, tarBall); err != nil {
			log.Error(err, "Error downloading asset", "name", name, "source", source, "tarBall", tarBall)
			return errors.Wrapf(err, "Error downloading asset %v; uri %v", name, source)
		}

		destDir := filepath.Join(m.config.GetAssetsDir(), name)
		// TODO(jeremy): Unpack the asset
		if err := unpackTarball(tarBall, destDir, a.stripPrefix); err != nil {
			return errors.Wrapf(err, "Error unpacking tarball %v", tarBall)
		}
	}
	return nil
}

// unpackTarball copies assets in the source tarbell matching the to the destination tarball
// glob is a glob pattern to match against the tarball
// strip is a path prefix to strip from all paths
// destPrefix is a path prefix to add to all paths
func unpackTarball(srcTarball string, dest string, stripPrefix string) error {
	log := zapr.NewLogger(zap.L())
	log.Info("Unpacking tarball", "srcTarball", srcTarball, "dest", dest, "stripPrefix", stripPrefix)
	factory := &files.Factory{}
	helper, err := factory.Get(srcTarball)
	if err != nil {
		return errors.Wrapf(err, "Error opening tarball %v", srcTarball)
	}
	reader, err := helper.NewReader(srcTarball)
	if err != nil {
		return errors.Wrapf(err, "Error opening tarball %v", srcTarball)
	}

	// Create a tar reader
	tarReader := tar.NewReader(reader)

	// Iterate over each file in the tarball
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			// Reached the end of the tarball
			return nil
		}

		if err != nil {
			return errors.Wrapf(err, "Error reading tar header:")
		}

		if header.Size == 0 {
			continue
		}

		path := header.Name
		if stripPrefix != "" {
			newPath, err := filepath.Rel(stripPrefix, header.Name)
			if err != nil {
				// Keep going
				log.Error(err, "Error stripping prefix", "prefix", stripPrefix, "path", header.Name)
			} else {
				path = newPath
			}
		}

		destPath := filepath.Join(dest, path)

		fileDir := filepath.Dir(destPath)
		if err := os.MkdirAll(fileDir, 0755); err != nil {
			return errors.Wrapf(err, "Error creating directory %v", fileDir)
		}

		log.V(logs.Debug).Info("Unpacking tarball entry", "header", header.Name, "size", header.Size, "file", destPath)
		f, err := os.Create(destPath)
		if err != nil {
			return errors.Wrapf(err, "Error creating file %v", destPath)
		}
		// Read the file contents
		_, err = io.CopyN(f, tarReader, header.Size)
		if err != nil {
			return errors.Wrapf(err, "Error reading file contents")
		}
		if err := f.Close(); err != nil {
			return errors.Wrapf(err, "Error closing file %v", destPath)
		}
	}
}

// resolveTag checks if repo has a tag and if not it adds the tag specified by tag
func resolveTag(repo string, tag string) (string, error) {
	ref, err := name.ParseReference(repo, name.WithDefaultTag(tag))
	if err != nil {
		return "", errors.Wrapf(err, "Error parsing reference %v", repo)
	}
	return ref.Name(), nil
}
