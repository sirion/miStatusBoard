package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Title           string                     `yaml:"title" json:"title"`
	Authorization   AuthorizationConfiguration `yaml:"authorization" json:"-"`
	RefreshInterval float64                    `yaml:"refreshInterval" json:"-"`
	Groups          []*Group                   `yaml:"groups" json:"groups"`
}

type AuthorizationConfiguration struct {
	Type            string   `yaml:"type" json:"-"`
	Header          string   `yaml:"header" json:"-"`
	Users           []string `yaml:"users" json:"-"`
	authorizedUsers map[string]bool
}

func ReadConfiguration(configPath string) (*Configuration, error) {
	config := Configuration{}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Quick access map for authorization checking
	config.Authorization.authorizedUsers = make(map[string]bool, len(config.Authorization.Users))
	for _, u := range config.Authorization.Users {
		config.Authorization.authorizedUsers[u] = true
	}

	if config.RefreshInterval < 10 {
		out("Configuration: RefreshInterval too low: %f. Set to 10\n", config.RefreshInterval)
		config.RefreshInterval = 10
	}

	return &config, nil
}
