package config

func SetAzureDeployment(cfg *Config, d AzureDeployment) {
	if cfg.AzureOpenAI == nil {
		cfg.AzureOpenAI = &AzureOpenAIConfig{}
	}
	if cfg.AzureOpenAI.Deployments == nil {
		cfg.AzureOpenAI.Deployments = make([]AzureDeployment, 0, 1)
	}
	// First check if there is a deployment for the model and if there is update it
	for i := range cfg.AzureOpenAI.Deployments {
		if cfg.AzureOpenAI.Deployments[i].Model == d.Model {
			cfg.AzureOpenAI.Deployments[i].Deployment = d.Deployment
			return
		}
	}

	cfg.AzureOpenAI.Deployments = append(cfg.AzureOpenAI.Deployments, d)
}
