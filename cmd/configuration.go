package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Title             string                     `yaml:"title" json:"title"`
	Authorization     AuthorizationConfiguration `yaml:"authorization" json:"-"`
	RefreshInterval   float64                    `yaml:"refreshInterval" json:"refresh_interval"`
	DefaultHttpMethod string                     `yaml:"default_http_method" json:"-"`
	Groups            []*Group                   `yaml:"groups" json:"groups"`
}

type AuthorizationConfiguration struct {
	Type            string   `yaml:"type" json:"-"`
	Header          string   `yaml:"header" json:"-"`
	Users           []string `yaml:"users" json:"-"`
	Cert            string   `yaml:"cert" json:"-"`
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

	if config.Authorization.Type == AUTH_TYPE_CERT && config.Authorization.Cert == "" {
		return nil, fmt.Errorf("authorization.cert must be set when authorization.type is \"%s\"", config.Authorization.Type)
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

	switch config.DefaultHttpMethod {
	case http.MethodGet, http.MethodHead:
		// Valid

	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		fmt.Fprintf(os.Stderr, "Error: Default HTTP Method %s not supported. Defaulting to GET\n", config.DefaultHttpMethod)
		fallthrough
	default:
		config.DefaultHttpMethod = http.MethodGet
	}

	for _, group := range config.Groups {
		for _, endpoint := range group.Endpoints {
			endpoint.Method = strings.ToUpper(endpoint.Method)

			// Validate HTTP Method
			switch endpoint.Method {
			case http.MethodGet, http.MethodHead:
				// Valid

			case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
				fmt.Fprintf(os.Stderr, "Error: HTTP Method %s for endpoint %s not supported. Defaulting to %s\n", endpoint.Method, endpoint.URL, config.DefaultHttpMethod)
				fallthrough
			default:
				endpoint.Method = config.DefaultHttpMethod
			}
		}
	}

	return &config, nil
}
