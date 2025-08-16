package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"yuki/internal/logger"
)

type Config struct {
	configPath string
	data       map[string]interface{}
}

var globalConfig *Config

func GetGlobalConfig() *Config {
	if globalConfig == nil {
		globalConfig = New()
	}
	return globalConfig
}

func New() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	
	configDir := filepath.Join(homeDir, ".yuki")
	configPath := filepath.Join(configDir, "config.json")
	
	config := &Config{
		configPath: configPath,
		data:       make(map[string]interface{}),
	}
	
	config.load()
	return config
}

func (c *Config) Get(key string) (interface{}, bool) {
	value, exists := c.data[key]
	return value, exists
}

func (c *Config) GetString(key string, defaultValue string) string {
	if value, exists := c.data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (c *Config) GetBool(key string, defaultValue bool) bool {
	if value, exists := c.data[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func (c *Config) Set(key string, value interface{}) error {
	c.data[key] = value
	return c.save()
}

func (c *Config) Delete(key string) error {
	delete(c.data, key)
	return c.save()
}

func (c *Config) List() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

func (c *Config) load() {
	configDir := filepath.Dir(c.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Debug("Failed to create config directory: %v", err)
		return
	}
	
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Debug("Failed to read config file: %v", err)
		}
		return
	}
	
	if err := json.Unmarshal(data, &c.data); err != nil {
		logger.Debug("Failed to parse config file: %v", err)
	}
}

func (c *Config) save() error {
	configDir := filepath.Dir(c.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(c.configPath, data, 0644)
}


func (c *Config) SetGitHubToken(token string) error {
	return c.Set("github_token", token)
}


func (c *Config) GetGitHubToken() string {
	
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token
	}
	
	
	return c.GetString("github_token", "")
}
