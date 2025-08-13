package utils

import (
	"api-gateway/internal/config"
	"fmt"
)

// Helper to get service configuration and apply global defaults
func GetServiceConfig(serviceName string, cfg *config.MainConfig) (*config.ServiceConfig, error) {
	service, exists := cfg.Services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Create a copy to avoid modifying the original
	serviceCopy := service

	// Only apply global settings if Global config exists
	if cfg.Global != nil {
		// Apply global settings if service-specific ones are not set
		if serviceCopy.RateLimit == nil && cfg.Global.RateLimit != nil {
			serviceCopy.RateLimit = cfg.Global.RateLimit
		}
		if len(serviceCopy.IPAllowList) == 0 && len(cfg.Global.IPAllowList) > 0 {
			serviceCopy.IPAllowList = cfg.Global.IPAllowList
		}
		if len(serviceCopy.IPBlockList) == 0 && len(cfg.Global.IPBlockList) > 0 {
			serviceCopy.IPBlockList = cfg.Global.IPBlockList
		}
		if len(serviceCopy.UserAgentAllowlist) == 0 && len(cfg.Global.UserAgentAllowlist) > 0 {
			serviceCopy.UserAgentAllowlist = cfg.Global.UserAgentAllowlist
		}
		if len(serviceCopy.UserAgentBlocklist) == 0 && len(cfg.Global.UserAgentBlocklist) > 0 {
			serviceCopy.UserAgentBlocklist = cfg.Global.UserAgentBlocklist
		}
	}

	return &serviceCopy, nil
}
