package config

import (
	"api-gateway/internal/constants"
	"api-gateway/pkg/config"
)

var manager *config.Manager

// InitConfig initializes the configuration manager
func InitConfig() error {
	var err error
	manager, err = config.NewManager("config/main.yaml")
	if err != nil {
		return err
	}
	return manager.Start()
}

// StopConfig stops the configuration manager
func StopConfig() {
	if manager != nil {
		manager.Stop()
	}
}

// GetConfig returns the current configuration
func GetConfig() constants.MainConfig {
	if manager == nil {
		return constants.MainConfig{}
	}
	return manager.GetConfig()
}
