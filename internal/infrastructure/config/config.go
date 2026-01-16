package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Scraper    ScraperConfig    `mapstructure:"scraper"`
	GRPC       GRPCConfig       `mapstructure:"grpc"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Algorithms AlgorithmConfig  `mapstructure:"algorithms"`
	Ensemble   EnsembleConfig   `mapstructure:"ensemble"`
	Backtest   BacktestConfig   `mapstructure:"backtest"`
}

// AppConfig represents application-level configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	LogLevel    string `mapstructure:"log_level"`
}

// ScraperConfig represents scraper configuration
type ScraperConfig struct {
	Vietlott VietlottScraperConfig `mapstructure:"vietlott"`
}

// VietlottScraperConfig represents Vietlott-specific scraper configuration
type VietlottScraperConfig struct {
	BaseURL    string        `mapstructure:"base_url"`
	Mega645Path  string      `mapstructure:"mega_645_path"`
	Power655Path string      `mapstructure:"power_655_path"`
	Timeout    time.Duration `mapstructure:"timeout"`
	RetryCount int           `mapstructure:"retry_count"`
	RateLimit  int           `mapstructure:"rate_limit"`
}

// GRPCConfig represents gRPC configuration
type GRPCConfig struct {
	TooPredict TooPredictGRPCConfig `mapstructure:"too_predict"`
	Server     ServerGRPCConfig     `mapstructure:"server"`
}

// TooPredictGRPCConfig represents gRPC client configuration for too_predict
type TooPredictGRPCConfig struct {
	Address    string        `mapstructure:"address"`
	Timeout    time.Duration `mapstructure:"timeout"`
	EnableTLS  bool          `mapstructure:"enable_tls"`
}

// ServerGRPCConfig represents gRPC server configuration
type ServerGRPCConfig struct {
	Port             int  `mapstructure:"port"`
	EnableReflection bool `mapstructure:"enable_reflection"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Type   string        `mapstructure:"type"` // "json" or "sqlite"
	SQLite SQLiteConfig  `mapstructure:"sqlite"`
	JSON   JSONConfig    `mapstructure:"json"`
}

// SQLiteConfig represents SQLite storage configuration
type SQLiteConfig struct {
	Path string `mapstructure:"path"`
}

// JSONConfig represents JSON file storage configuration
type JSONConfig struct {
	BasePath string `mapstructure:"base_path"`
}

// AlgorithmConfig represents algorithm configuration
type AlgorithmConfig struct {
	Enabled []string                    `mapstructure:"enabled"`
	Configs map[string]AlgorithmDetails `mapstructure:",remain"`
}

// AlgorithmDetails represents individual algorithm configuration
type AlgorithmDetails struct {
	Weight float64 `mapstructure:"weight"`
	// Add more algorithm-specific settings as needed
}

// EnsembleConfig represents ensemble configuration
type EnsembleConfig struct {
	VotingStrategy string  `mapstructure:"voting_strategy"` // "weighted", "majority", "confidence_weighted"
	MinPredictions int     `mapstructure:"min_predictions"`
}

// BacktestConfig represents backtesting configuration
type BacktestConfig struct {
	DefaultTestPeriodDays  int `mapstructure:"default_test_period_days"`
	DefaultTestPeriodDraws int `mapstructure:"default_test_period_draws"`
	EnableAutoWeightUpdate bool `mapstructure:"enable_auto_weight_update"`
}

// Load loads configuration from a file
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Set defaults
	setDefaults()

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("TOOL_PREDICT")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// LoadWithDefaults loads configuration with custom defaults
func LoadWithDefaults(configPath string, defaults map[string]interface{}) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Set custom defaults
	for key, value := range defaults {
		viper.SetDefault(key, value)
	}

	// Set standard defaults
	setDefaults()

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("TOOL_PREDICT")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("app.name", "tool_predict")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.log_level", "info")

	viper.SetDefault("scraper.vietlott.base_url", "https://vietlott.vn")
	viper.SetDefault("scraper.vietlott.timeout", 30*time.Second)
	viper.SetDefault("scraper.vietlott.retry_count", 3)
	viper.SetDefault("scraper.vietlott.rate_limit", 2)

	viper.SetDefault("grpc.too_predict.address", "localhost:50051")
	viper.SetDefault("grpc.too_predict.timeout", 10*time.Second)
	viper.SetDefault("grpc.too_predict.enable_tls", false)

	viper.SetDefault("grpc.server.port", 50052)
	viper.SetDefault("grpc.server.enable_reflection", true)

	viper.SetDefault("storage.type", "json")
	viper.SetDefault("storage.json.base_path", "./data")

	viper.SetDefault("ensemble.voting_strategy", "weighted")
	viper.SetDefault("ensemble.min_predictions", 2)

	viper.SetDefault("backtest.default_test_period_days", 30)
	viper.SetDefault("backtest.default_test_period_draws", 30)
	viper.SetDefault("backtest.enable_auto_weight_update", true)
}

// GetAlgorithmWeight returns the weight for a specific algorithm
func (c *Config) GetAlgorithmWeight(algorithmName string) float64 {
	if algoConfig, exists := c.Algorithms.Configs[algorithmName]; exists {
		return algoConfig.Weight
	}
	return 1.0 // default weight
}

// IsAlgorithmEnabled checks if an algorithm is enabled
func (c *Config) IsAlgorithmEnabled(algorithmName string) bool {
	for _, enabled := range c.Algorithms.Enabled {
		if enabled == algorithmName {
			return true
		}
	}
	return false
}
