package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlewi/foyle/app/api"

	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// TODO(jeremy): We should finish moving the configuration datastructure into the API package.
// However, we should keep the API package free of other dependencies (e.g. Cobra) so that might necessitate
// refactoring the code a bit more.

// Note: The application uses viper for configuration management. Viper merges configurations from various sources
//such as files, environment variables, and command line flags. After merging, viper unmarshals the configuration into the Configuration struct, which is then used throughout the application.

const (
	ConfigFlagName    = "config"
	LevelFlagName     = "level"
	appName           = "foyle"
	ConfigDir         = "." + appName
	defaultMaxResults = 3

	// defaultHTTPPort should be kept in sync with the default in RunMe
	// https://github.com/stateful/vscode-runme/blob/f1cc965ab0c4cdffa9adb70922e2da792d7e23de/package.json#L849
	// The value isn't 8080 because 8080 is over used and likely to conflict with other locally running services.
	defaultHTTPPort = 8877
)

// Config represents the persistent configuration data for Foyle.
//
// Currently, the format of the data on disk and in memory is identical. In the future, we may modify this to simplify
// changes to the disk format and to store in-memory values that should not be written to disk.
type Config struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion" yamltags:"required"`
	Kind       string `json:"kind" yaml:"kind" yamltags:"required"`

	Logging Logging          `json:"logging" yaml:"logging"`
	Server  ServerConfig     `json:"server" yaml:"server"`
	Assets  *AssetConfig     `json:"assets,omitempty" yaml:"assets,omitempty"`
	Agent   *api.AgentConfig `json:"agent,omitempty" yaml:"agent,omitempty"`
	OpenAI  *OpenAIConfig    `json:"openai,omitempty" yaml:"openai,omitempty"`
	// AzureOpenAI contains configuration for Azure OpenAI. A non nil value means use Azure OpenAI.
	AzureOpenAI *AzureOpenAIConfig `json:"azureOpenAI,omitempty" yaml:"azureOpenAI,omitempty"`

	Telemetry *TelemetryConfig `json:"telemetry,omitempty" yaml:"telemetry,omitempty"`

	// TODO(jeremy): Should we move this into the experiment?
	Eval *EvalConfig `json:"eval,omitempty" yaml:"eval,omitempty"`

	Learner *LearnerConfig `json:"learner,omitempty" yaml:"learner,omitempty"`

	Replicate *ReplicateConfig `json:"replicate,omitempty" yaml:"replicate,omitempty"`
	Anthropic *AnthropicConfig `json:"anthropic,omitempty" yaml:"anthropic,omitempty"`
}

type LearnerConfig struct {
	// LogDirs is an additional list of directories to search for logs.
	// Deprecated: We should remove this in v1alpha2. This is no longer needed now that we no longer rely on processing
	// RunMe's logs but rather the UI sends the logs directly to the server.
	LogDirs []string `json:"logDirs" yaml:"logDirs"`

	// ExampleDirs is the list of directories to read/write examples.
	// Can be a local path or GCS URI.
	ExampleDirs []string `json:"exampleDirs" yaml:"exampleDirs"`
}

type EvalConfig struct {
	// GCPServiceAccount is the service account to use to update Google Sheets
	GCPServiceAccount string `json:"gcpServiceAccount" yaml:"gcpServiceAccount"`
}

// ServerConfig configures the server
type ServerConfig struct {
	// BindAddress is the address to bind to. Default is 0.0.0.0
	BindAddress string `json:"bindAddress" yaml:"bindAddress"`

	// HttpPort is the port for the http service
	HttpPort int `json:"httpPort" yaml:"httpPort"`

	// GRPCPort is the port for the gRPC service
	// Deprecated: GRPCPort is no longer used and should be removed in the next version of the config.
	GRPCPort int `json:"grpcPort,omitempty" yaml:"grpcPort,omitempty"`

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

type AnthropicConfig struct {
	// APIKeyFile is the path to the file containing the API key
	APIKeyFile string `json:"apiKeyFile" yaml:"apiKeyFile"`
}

type ReplicateConfig struct {
	// APIKeyFile is the path to the file containing the API key
	APIKeyFile string `json:"apiKeyFile" yaml:"apiKeyFile"`
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
	VSCode         *Asset `json:"vsCode,omitempty" yaml:"vsCode,omitempty"`
	FoyleExtension *Asset `json:"foyleExtension,omitempty" yaml:"foyleExtension,omitempty"`
}

type Asset struct {
	// URI is the URI of the source for the asset
	URI string `json:"uri" yaml:"uri"`
}

type Logging struct {
	Level  string `json:"level,omitempty" yaml:"level,omitempty"`
	LogDir string `json:"logDir,omitempty" yaml:"logDir,omitempty"`
	// Sinks is a list of sinks to write logs to.
	// Use stderr to write to stderr.
	// Use gcplogs:///projects/${PROJECT}/logs/${LOGNAME} to write to Google Cloud Logging
	Sinks []LogSink `json:"sinks,omitempty" yaml:"sinks,omitempty"`
}

type LogSink struct {
	// Set to true to write logs in JSON format
	JSON bool `json:"json,omitempty" yaml:"json,omitempty"`
	// Path is the path to write logs to. Use "stderr" to write to stderr.
	// Use gcplogs:///projects/${PROJECT}/logs/${LOGNAME} to write to Google Cloud Logging
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
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

func (c *Config) GetLogDir() string {
	if c.Logging.LogDir != "" {
		return c.Logging.LogDir
	}

	return filepath.Join(c.GetConfigDir(), "logs")
}

func (c *Config) GetLogOffsetsFile() string {
	return filepath.Join(c.GetLogDir(), "offsets.v1.json")
}

func (c *Config) GetRawLogDir() string {
	return filepath.Join(c.GetLogDir(), "raw")
}

func (c *Config) GetLogEntriesDBDir() string {
	return filepath.Join(c.GetLogDir(), "logEntries")
}

func (c *Config) GetBlocksDBDir() string {
	return filepath.Join(c.GetLogDir(), "blocks")
}

func (c *Config) GetTracesDBDir() string {
	return filepath.Join(c.GetLogDir(), "traces")
}

func (c *Config) GetSessionsDB() string {
	return filepath.Join(c.GetLogDir(), "sessions.sqllite3")
}

func (c *Config) GetTrainingDirs() []string {
	if c.Learner == nil {
		return []string{}
	}

	dirs := c.Learner.ExampleDirs
	// If dirs isn't set default to a local training directory
	if len(dirs) == 0 {
		dirs = []string{filepath.Join(c.GetConfigDir(), "training")}
	}
	return dirs
}

func (c *Config) GetLogLevel() string {
	if c.Logging.Level == "" {
		return "info"
	}
	return c.Logging.Level
}

// GetConfigDir returns the configuration directory
func (c *Config) GetConfigDir() string {
	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		return filepath.Dir(configFile)
	}

	// Since there is no config file we will use the default config directory.
	return binHome()
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

func (c *Config) UseRAG() bool {
	if c.Agent == nil || c.Agent.RAG == nil {
		return false
	}
	return c.Agent.RAG.Enabled
}

func (c *Config) RagMaxResults() int {
	if c.Agent == nil || c.Agent.RAG == nil {
		return -1
	}
	if c.Agent.RAG.MaxResults <= 0 {
		return defaultMaxResults
	}
	return c.Agent.RAG.MaxResults
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

func (c *Config) EvalMode() bool {
	if c.Agent == nil {
		return false
	}
	return c.Agent.EvalMode
}

// DeepCopy returns a deep copy.
func (c *Config) DeepCopy() Config {
	b, err := json.Marshal(c)
	if err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to marshal config")
		panic(err)
	}
	var copy Config
	if err := json.Unmarshal(b, &copy); err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to unmarshal config")
		panic(err)
	}
	return copy
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

func (c *Config) APIPrefix() string {
	// N.B. don't include leading or trailing slashes in the prefix because the code in server.go assumes there isn't any
	return "api"
}

// APIBaseURL returns the base URL for the API
func (c *Config) APIBaseURL() string {
	return fmt.Sprintf("http://%s:%d/%s", c.Server.BindAddress, c.Server.HttpPort, c.APIPrefix())
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
	viper.SetDefault("server.httpPort", defaultHTTPPort)
	// gRPC typically uses 50051. If we use that as the default we might end up conflicting with other gRPC services
	// running by default.
	viper.SetDefault("server.grpcPort", 9080)

	// See https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts
	// If we start using really slow models we may need to bump these to avoid timeouts.
	viper.SetDefault("server.httpMaxWriteTimeout", 1*time.Minute)
	viper.SetDefault("server.httpMaxReadTimeout", 1*time.Minute)
}

func setAgentDefaults() {
	viper.SetDefault("agent.model", DefaultModel)
}

func DefaultConfigFile() string {
	return binHome() + "/config.yaml"
}

// NewWithTempDir initializes an empty configuration in a temporary directory.
// It is intended solely for use in tests where we want to use a temporary directory.
func NewWithTempDir() (*Config, error) {
	dir, err := os.MkdirTemp("", "foyleConfig")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create temporary directory")
	}
	return &Config{
		APIVersion: "v1alpha1",
		Kind:       "Config",
		Logging: Logging{
			LogDir: dir,
		},
	}, nil
}
