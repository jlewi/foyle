package replicate

import (
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/monogo/files"
	"github.com/pkg/errors"
	repGo "github.com/replicate/replicate-go"
	"strings"
)

func NewClient(cfg config.Config) (*repGo.Client, error) {
	if cfg.Replicate == nil {
		return nil, errors.New("Replicate config is nil; You must configure the replicate model provider to create a replicate client")
	}
	token, err := files.Read(cfg.Replicate.APIKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read Replicate API token from file %s", cfg.Replicate.APIKeyFile)
	}
	r8, err := repGo.NewClient(repGo.WithToken(strings.TrimSpace(string(token))))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create Replicate client")
	}
	return r8, nil
}
