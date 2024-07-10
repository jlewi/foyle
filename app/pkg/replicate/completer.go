package replicate

import (
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

const (
	defaultModel = "meta/meta-llama-3-8b-instruct"
)

func NewCompleter(cfg config.Config, client *openai.Client) (*oai.Completer, error) {
	log := zapr.NewLogger(zap.L())
	if cfg.Agent == nil {
		cfg.Agent = &api.AgentConfig{}
	}
	if cfg.Agent == nil || cfg.Agent.Model == "" {

		log.Info("No model specified; using default model", "model", defaultModel)
		cfg.Agent.Model = "meta/meta-llama-3-8b-instruct"
	}
	return oai.NewCompleter(cfg, client)
}
