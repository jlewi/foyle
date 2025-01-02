package oai

import (
	"context"
	"github.com/jlewi/foyle/app/api/oaiapi"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"strings"
)

// FileSyncer is a controller to sync files to an assistant.
// https://platform.openai.com/docs/api-reference/files
type FileSyncer struct {
	cfg    config.Config
	client *openai.Client
}

// NewFileSyncer creates a new controller for OpenAI file sync
func NewFileSyncer(cfg config.Config) (*FileSyncer, error) {
	// Delay creation of the client so that we don't create it just to register the controller
	return &FileSyncer{cfg: cfg}, nil
}

// ReconcileNode reconciles the state of the resource.
func (f *FileSyncer) ReconcileNode(ctx context.Context, node *yaml.RNode) error {
	s := &oaiapi.FileSync{}
	if err := node.YNode().Decode(s); err != nil {
		return errors.Wrap(err, "Failed to decode FileSync")
	}

	return f.Apply(ctx, s)
}

func (f *FileSyncer) Apply(ctx context.Context, s *oaiapi.FileSync) error {
	log := logs.FromContext(ctx)
	if f.client == nil {
		client, err := NewClient(f.cfg)
		if err != nil {
			return errors.Wrap(err, "Failed to create OpenAI client")
		}
		f.client = client
	}
	// TODO(jeremy): We shouldn't assume we only want to match markdown files.
	// How can we support a suitable glob like syntax?
	mdFiles, err := findMarkdownFiles(s.Spec.Source)
	if err != nil {
		errors.Wrapf(err, "Failed to find markdown files in %v", s.Spec.Source)
	}

	client := f.client

	files, err := client.ListFiles(ctx)
	if err != nil {
		errors.Wrap(err, "Failed to list files")
	}

	alreadyUploaded := make(map[string]string)
	for _, f := range files.Files {
		alreadyUploaded[f.FileName] = f.ID
	}

	fileIDs := make([]string, 0, len(mdFiles))

	for _, mdFile := range mdFiles {
		relPath, err := filepath.Rel(s.Spec.Source, mdFile)
		if err != nil {
			errors.Wrapf(err, "Failed to get relative path for %v", mdFile)
		}

		if fid, ok := alreadyUploaded[relPath]; ok {
			log.Info("File already uploaded", "file", relPath, "id", fid)
			fileIDs = append(fileIDs, fid)
			continue
		}

		req := &openai.FileRequest{
			FileName: relPath,
			FilePath: mdFile,
			Purpose:  string(openai.PurposeAssistants),
		}
		newFile, err := f.client.CreateFile(ctx, *req)
		if err != nil {
			errors.Wrapf(err, "Failed to create file %v", mdFile)
		}
		log.Info("Uploaded file", "file", mdFile, "id", newFile.ID)
		fileIDs = append(fileIDs, newFile.ID)
	}

	if err != nil {
		errors.Wrap(err, "Failed to list files")
	}

	req := &openai.VectorStoreFileBatchRequest{
		FileIDs: fileIDs,
	}
	log.Info("Creating vector store file batch", "numFileIDs", len(fileIDs))
	if _, err := client.CreateVectorStoreFileBatch(ctx, s.Spec.VectorStoreID, *req); err != nil {
		errors.Wrapf(err, "Failed to create vector store file batch")
	}

	return nil
}

func findMarkdownFiles(dir string) ([]string, error) {
	var markdownFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			markdownFiles = append(markdownFiles, path)
		}
		return nil
	})

	return markdownFiles, err
}
