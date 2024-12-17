package constants

import "time"

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
