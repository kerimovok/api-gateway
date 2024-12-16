package config

import (
	"api-gateway/internal/constants"
	"api-gateway/pkg/utils"
	"fmt"
	"os"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Manager struct {
	config   constants.MainConfig
	mu       sync.RWMutex
	watcher  *Watcher
	filePath string
}

// NewManager creates a new config manager
func NewManager(configPath string) (*Manager, error) {
	m := &Manager{
		filePath: configPath,
	}

	if err := m.loadConfig(); err != nil {
		return nil, err
	}

	utils.LogInfo("Config loaded successfully from " + configPath)

	// Set up config watcher
	watcher, err := New(configPath, m.loadConfig)
	if err != nil {
		return nil, err
	}

	m.watcher = watcher
	return m, nil
}

// Start begins watching for config changes
func (m *Manager) Start() error {
	if err := m.watcher.Start(); err != nil {
		utils.LogError("Failed to start config watcher", err)
		return err
	}
	utils.LogInfo("Config watcher started successfully")
	return nil
}

// Stop stops the config manager and watcher
func (m *Manager) Stop() {
	if m.watcher != nil {
		m.watcher.Stop()
		utils.LogInfo("Config watcher stopped")
	}
}

// GetConfig returns a thread-safe copy of the config
func (m *Manager) GetConfig() constants.MainConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// validateConfig checks if the loaded config is valid
func (m *Manager) validateConfig(cfg *constants.MainConfig) error {
	validate := validator.New()

	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}

// loadConfig handles the actual loading of config files
func (m *Manager) loadConfig() error {
	if err := godotenv.Load(); err != nil {
		if os.Getenv("GO_ENV") != "production" {
			utils.LogWarn("Failed to load .env file")
		}
	}

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	var newConfig constants.MainConfig
	if err := yaml.Unmarshal(data, &newConfig); err != nil {
		return err
	}

	// Validate config before applying
	if err := m.validateConfig(&newConfig); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Assign service names from keys in the map
	for name, service := range newConfig.Services {
		service.Name = name
		newConfig.Services[name] = service
	}

	m.mu.Lock()
	m.config = newConfig
	m.mu.Unlock()

	return nil
}
