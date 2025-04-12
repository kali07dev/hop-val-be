package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

type ExternalAPIConfig struct { // New struct
	PropertiesURL string `yaml:"properties_url"`
}

type Config struct {
	Database    DatabaseConfig    `yaml:"database"`
	ExternalAPI ExternalAPIConfig `yaml:"external_api"` 
}

var Cfg *Config 

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	Cfg = &config 
	fmt.Println("Configuration loaded successfully.")
	return nil
}

// Helper function to get the loaded config 
func GetConfig() *Config {
	if Cfg == nil {
		// Handle case where config hasn't been loaded, maybe panic or return error
		panic("Configuration not loaded!")
	}
	return Cfg
}