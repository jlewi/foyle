package config

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Note: The application uses viper for configuration management. Viper merges configurations from various sources
//such as files, environment variables, and command line flags. After merging, viper unmarshals the configuration into the Configuration struct, which is then used throughout the application.

const (
	ConfigFlagName = "config"
	LevelFlagName  = "level"
	appName        = "foyle"
	ConfigDir      = "." + appName

	defaultVSCodeImage = "ghcr.io/jlewi/vscode-web-assets:latest"
	defaultFoyleImage  = "ghcr.io/jlewi/foyle-vscode-ext:latest"
)

// Config represents the persistent configuration data for Foyle.
//
// Currently, the format of the data on disk and in memory is identical. In the future, we may modify this to simplify
// changes to the disk format and to store in-memory values that should not be written to disk.
type Config struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion" yamltags:"required"`
	Kind       string `json:"kind" yaml:"kind" yamltags:"required"`

	Logging Logging       `json:"logging" yaml:"logging"`
	Server  ServerConfig  `json:"server" yaml:"server"`
	Assets  AssetConfig   `json:"assets" yaml:"assets"`
	Agent   *AgentConfig  `json:"agent,omitempty" yaml:"agent,omitempty"`
	OpenAI  *OpenAIConfig `json:"openai,omitempty" yaml:"openai,omitempty"`
	// AzureOpenAI contains configuration for Azure OpenAI. A non nil value means use Azure OpenAI.
	AzureOpenAI *AzureOpenAIConfig `json:"azureOpenAI,omitempty" yaml:"azureOpenAI,omitempty"`

	Telemetry *TelemetryConfig `json:"telemetry,omitempty" yaml:"telemetry,omitempty"`
}

type AgentConfig struct {
	// Model is the name of the model to use to generate completions
	Model string `json:"model" yaml:"model"`
}

// ServerConfig configures the server
type ServerConfig struct {
	// BindAddress is the address to bind to. Default is 0.0.0.0
	BindAddress string `json:"bindAddress" yaml:"bindAddress"`

	// HttpPort is the port for the http service
	HttpPort int `json:"httpPort" yaml:"httpPort"`

	// GRPCPort is the port for the gRPC service
	GRPCPort int `json:"grpcPort" yaml:"grpcPort"`

	// CORS contains the CORS configuration
	CORS *CorsConfig `json:"cors,omitempty" yaml:"cors,omitempty"`

	// HttpMaxReadTimeout is the max read duration.
	// Ref: https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts
	HttpMaxReadTimeout time.Duration `json:"httpMaxReadTimeout" yaml:"httpMaxReadTimeout"`

	// HttpMaxWriteTimeout is the max write duration.
	HttpMaxWriteTimeout time.Duration `json:"httpMaxWriteTimeout" yaml:"httpMaxWriteTimeout"`
}

type OpenAIConfig struct {
	// APIKeyFile is the path to the file containing the API key
	APIKeyFile string `json:"apiKeyFile" yaml:"apiKeyFile"`

	// BaseURL is the baseURL for the API.
	BaseURL string `json:"baseURL" yaml:"baseURL"`
}

type AzureOpenAIConfig struct {
	// APIKeyFile is the path to the file containing the API key
	APIKeyFile string `json:"apiKeyFile" yaml:"apiKeyFile"`

	// BaseURL is the baseURL for the API.
	// This can be obtained using the Azure CLI with the command:
	// az cognitiveservices account show \
	//    --name <myResourceName> \
	//    --resource-group  <myResourceGroupName> \
	//    | jq -r .properties.endpoint
	BaseURL string `json:"baseURL" yaml:"baseURL"`

	// Deployments is a list of Azure deployments of various models.
	Deployments []AzureDeployment `json:"deployments" yaml:"deployments"`
}

type AzureDeployment struct {
	// Deployment is the Azure Deployment name
	Deployment string `json:"deployment" yaml:"deployment"`

	// Model is the OpenAI name for this model
	// This is used to map OpenAI models to Azure deployments
	Model string `json:"model" yaml:"model"`
}

type CorsConfig struct {
	// AllowedOrigins is a list of origins allowed to make cross-origin requests.
	AllowedOrigins []string `json:"allowedOrigins" yaml:"allowedOrigins"`
	// AllowedHeaders is a list of headers which are allowed
	AllowedHeaders []string `json:"allowedHeaders" yaml:"allowedHeaders"`

	// VSCodeTestServerPort is the port of the VSCode test server
	// This will be a value like http://localhost:3000. This enables special CORS handling because
	// the vscode-test-web server generates a random prefix so the server name will be something like
	// http://v--19cf5ppcsqee9rrkfifq1ajc8b7nv2t96593a6n6bn95st339ul8.localhost:3000
	// Setting VSCodeTestServerPort origin will allow the server to accept requests from the test server. It should
	// Only be used during development
	VSCodeTestServerPort *int `json:"vsCodeTestServerPort" yaml:"vsCodeTestServerPort"`
}

// AssetConfig configures the assets
type AssetConfig struct {
	VSCode         Asset `json:"vsCode" yaml:"vsCode"`
	FoyleExtension Asset `json:"foyleExtension" yaml:"foyleExtension"`
}

type Asset struct {
	// URI is the URI of the source for the asset
	URI string `json:"uri" yaml:"uri"`
}

type Logging struct {
	Level string `json:"level" yaml:"level"`
}

type TelemetryConfig struct {
	Honeycomb *HoneycombConfig `json:"honeycomb,omitempty" yaml:"honeycomb,omitempty"`
}

type HoneycombConfig struct {
	// APIKeyFile is the Honeycomb API key
	APIKeyFile string `json:"apiKeyFile" yaml:"apiKeyFile"`
}

func (c *Config) GetModel() string {
	if c.Agent == nil || c.Agent.Model == "" {
		return DefaultModel
	}

	return c.Agent.Model
}
func (c *Config) GetLogLevel() string {
	if c.Logging.Level == "" {
		return "info"
	}
	return c.Logging.Level
}

// GetConfigDir returns the configuration directory
func (c *Config) GetConfigDir() string {
	return filepath.Dir(viper.ConfigFileUsed())
}

// IsValid validates the configuration and returns any errors.
func (c *Config) IsValid() []string {
	problems := make([]string, 0, 1)
	return problems
}

// GetAssetsDir returns the directory where assets are stored.
func (c *Config) GetAssetsDir() string {
	// TODO(jeremy): Should we make this configurable?
	return filepath.Join(c.GetConfigDir(), "assets")
}

func (c *Config) UseHoneycomb() bool {
	if c.Telemetry == nil {
		return false
	}
	if c.Telemetry.Honeycomb == nil {
		return false
	}
	if c.Telemetry.Honeycomb.APIKeyFile == "" {
		return false
	}
	return true
}

// InitViper function is responsible for reading the configuration file and environment variables, if they are set.
// The results are stored in viper. To retrieve a configuration, use the GetConfig function.
// The function accepts a cmd parameter which allows binding to command flags.
func InitViper(cmd *cobra.Command) error {
	// Ref https://github.com/spf13/viper#establishing-defaults
	viper.SetEnvPrefix(appName)
	// name of config file (without extension)
	viper.SetConfigName("config")
	// make home directory the first search path
	viper.AddConfigPath("$HOME/." + appName)

	// Without the replacer overriding with environment variables doesn't work
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	setAgentDefaults()
	setServerDefaults()
	setAssetDefaults()

	// We need to attach to the command line flag if it was specified.
	keyToflagName := map[string]string{
		ConfigFlagName:             ConfigFlagName,
		"logging." + LevelFlagName: LevelFlagName,
	}

	if cmd != nil {
		for key, flag := range keyToflagName {
			if err := viper.BindPFlag(key, cmd.Flags().Lookup(flag)); err != nil {
				return err
			}
		}
	}

	// Ensure the path for the config file path is set
	// Required since we use viper to persist the location of the config file so can save to it.
	cfgFile := viper.GetString(ConfigFlagName)
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log := zapr.NewLogger(zap.L())
			log.Error(err, "config file not found", "file", cfgFile)
			return nil
		}
		if _, ok := err.(*fs.PathError); ok {
			log := zapr.NewLogger(zap.L())
			log.Error(err, "config file not found", "file", cfgFile)
			return nil
		}
		return err
	}
	return nil
}

// GetConfig returns a configuration created from the viper configuration.
func GetConfig() *Config {
	// We do this as a way to load the configuration while still allowing values to be overwritten by viper
	cfg := &Config{}

	if err := viper.Unmarshal(cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal configuration; error %v", err))
	}

	return cfg
}

func binHome() string {
	log := zapr.NewLogger(zap.L())
	usr, err := user.Current()
	homeDir := ""
	if err != nil {
		log.Error(err, "failed to get current user; falling back to temporary directory for homeDir", "homeDir", os.TempDir())
		homeDir = os.TempDir()
	} else {
		homeDir = usr.HomeDir
	}
	p := filepath.Join(homeDir, ConfigDir)

	return p
}

// Write saves the configuration to a file.
func (c *Config) Write(cfgFile string) error {
	log := zapr.NewLogger(zap.L())
	if cfgFile == "" {
		return errors.Errorf("no config file specified")
	}
	configDir := filepath.Dir(cfgFile)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Info("creating config directory", "dir", configDir)
		if err := os.Mkdir(configDir, 0700); err != nil {
			return errors.Wrapf(err, "Ffailed to create config directory %s", configDir)
		}
	}

	f, err := os.Create(cfgFile)
	if err != nil {
		return err
	}

	return yaml.NewEncoder(f).Encode(c)
}

func setServerDefaults() {
	viper.SetDefault("server.bindAddress", "0.0.0.0")
	viper.SetDefault("server.httpPort", 8080)
	// gRPC typically uses 50051. If we use that as the default we might end up conflicting with other gRPC services
	// running by default.
	viper.SetDefault("server.grpcPort", 9080)

	// See https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts
	// If we start using really slow models we may need to bump these to avoid timeouts.
	viper.SetDefault("server.httpMaxWriteTimeout", 1*time.Minute)
	viper.SetDefault("server.httpMaxReadTimeout", 1*time.Minute)
}

func setAssetDefaults() {
	viper.SetDefault("assets.vsCode.uri", defaultVSCodeImage)
	viper.SetDefault("assets.foyleExtension.uri", defaultFoyleImage)
}

func setAgentDefaults() {
	viper.SetDefault("agent.model", DefaultModel)
}

func DefaultConfigFile() string {
	return binHome() + "/config.yaml"
}
