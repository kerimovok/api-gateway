package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	pkgConfig "github.com/kerimovok/go-pkg-utils/config"
	"gopkg.in/yaml.v3"
)

// Struct for rate limit settings
type RateLimitConfig struct {
	Enabled     *bool         `yaml:"enabled"`
	MaxRequests int           `yaml:"max_requests" validate:"required_if=Enabled true,gt=0"`
	Duration    time.Duration `yaml:"duration" validate:"required_if=Enabled true,gt=0"`
}

// Struct for service-specific settings
type AuthConfig struct {
	Enabled *bool  `yaml:"enabled"`
	Key     string `yaml:"key" validate:"required_if=Enabled true"`
	Value   string `yaml:"value" validate:"required_if=Enabled true"`
}

// FirewallConfig contains common configuration fields
type FirewallConfig struct {
	IPAllowList        []string `yaml:"ip_allowlist" validate:"omitempty,dive,ip|cidr"`
	IPBlockList        []string `yaml:"ip_blocklist" validate:"omitempty,dive,ip|cidr"`
	UserAgentAllowlist []string `yaml:"user_agent_allowlist" validate:"omitempty"`
	UserAgentBlocklist []string `yaml:"user_agent_blocklist" validate:"omitempty"`
}

// CacheConfig defines caching behavior
type CacheConfig struct {
	Enabled  *bool         `yaml:"enabled"`
	Duration time.Duration `yaml:"duration" validate:"required_if=Enabled true,gt=0"`
}

// ServiceConfig extends BaseConfig with service-specific settings
type ServiceConfig struct {
	FirewallConfig `yaml:",inline"`
	Name           string           `yaml:"name"`
	URL            string           `yaml:"url" validate:"required,url"`
	Auth           *AuthConfig      `yaml:"auth"`
	RateLimit      *RateLimitConfig `yaml:"rate_limit"`
	Cache          *CacheConfig     `yaml:"cache"`
}

// GlobalConfig extends BaseConfig with global settings
type GlobalConfig struct {
	FirewallConfig `yaml:",inline"`
	Logging        *bool            `yaml:"logging"`
	Cache          *CacheConfig     `yaml:"cache"`
	RateLimit      *RateLimitConfig `yaml:"rate_limit"`
}

// Root configuration struct
type MainConfig struct {
	Services map[string]ServiceConfig `yaml:"services" validate:"required,dive"`
	Global   *GlobalConfig            `yaml:"global"`
}

var (
	Main MainConfig
)

// LoadConfig loads the main configuration from the specified path
func LoadConfig() error {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		if pkgConfig.GetEnv("GO_ENV") != "production" {
			log.Println("Warning: Failed to load .env file")
		}
	}

	// Read config file
	data, err := os.ReadFile("config/main.yaml")
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config MainConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Assign service names from keys in the map
	for name, service := range config.Services {
		service.Name = name
		config.Services[name] = service
	}

	// Store config globally
	Main = config

	log.Println("Config loaded successfully from config/main.yaml")
	return nil
}

// GetConfig returns the current configuration
func GetConfig() MainConfig {
	return Main
}
