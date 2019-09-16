package config

import (
	"sync"

	"github.com/ITResourcesOSS/sgul"
)

type (
	// Cors defines the cors allowed resources struct.
	Cors struct {
		Origin  []string
		Methods []string
		Headers []string
	}

	// Endpoint defines an http endpoint for the Gateway.
	Endpoint struct {
		Schema   string
		Path     string
		Port     int
		Security struct {
			Enabled bool
			JWT     struct {
				Secret     string
				Expiration struct {
					Enabled bool
					Minutes int64
				}
			}
		}
	}

	// Gateway defines the full gateway configuration structure.
	Gateway struct {
		Endpoint Endpoint
		Admin    Endpoint
		Cors     Cors
	}

	// BalancingStrategy defines the proxy load balancing strategy.
	BalancingStrategy struct {
		Strategy string
	}

	// Proxy defines targets and balancing for an api proxy.
	Proxy struct {
		Path      string
		Schema    string
		Targets   []string
		Balancing BalancingStrategy
	}

	// APIEndpoint defines the API definition struct.
	APIEndpoint struct {
		Name    string
		Path    string
		Version string
		Proxy   Proxy
	}

	// API defines the API Definitions main configuration struct.
	API struct {
		Name      string
		Endpoints []APIEndpoint
	}

	// Configuration is the main configuration struct.
	Configuration struct {
		Service sgul.Service
		Gateway Gateway
		API     API
		Log     sgul.Log
	}
)

// Config is the global configuration singleton.
var Config *Configuration

// LoadConfiguration .
func LoadConfiguration() {
	var once sync.Once
	once.Do(func() {
		sgul.LoadConfiguration(&Config)
	})
}
