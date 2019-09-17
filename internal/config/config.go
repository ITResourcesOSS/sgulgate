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
		Endpoint  Endpoint
		Admin     Endpoint
		Cors      Cors
		Transport struct {
			MaxIdleConnections  int
			MaxIdleConnsPerHost int
			// IdleConnTimeout sets timeout in seconds for idle connections.
			IdleConnTimeout int
			// TLSHandshakeTimeout specifies the maximum amount of time in seconds waiting to
			// wait for a TLS handshake. Zero means no timeout.
			TLSHandshakeTimeout int
			// ExpectContinueTimeout, if non-zero, specifies the amount of
			// time to wait for a server's first response headers after fully
			// writing the request headers if the request has an
			// "Expect: 100-continue" header. Zero means no timeout and
			// causes the body to be sent immediately, without
			// waiting for the server to approve.
			// This time does not include the time to send the request header.
			ExpectContinueTimeout int
			DisableKeepAlives     bool
			DisableCompression    bool
			ResponseHeaderTimeout int
		}
		Dial struct {
			UpstreamTimeout int
			KeepAlive       int
			DualStack       bool
		}
	}

	// BalancingStrategy defines the proxy load balancing strategy.
	BalancingStrategy struct {
		Strategy string
	}

	// Proxy defines targets and balancing for an api proxy.
	Proxy struct {
		Path       string
		Schema     string
		Targets    []string
		MaxRetries int
		Balancing  BalancingStrategy
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
