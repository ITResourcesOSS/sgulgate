package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/itross/sgulgate/internal/config"
)

// ErrInvalidService returned if there is no api definition for service/version.
var ErrInvalidService = errors.New("invalid service/version")

// LoadBalance is the load balancing startegy implementation.
var LoadBalance = loadBalance

type transport struct {
	apiDef apiDefinition
	tr     *http.Transport
}

// Proxy .
type Proxy interface {
	Handler(w http.ResponseWriter, req *http.Request)
}

// APIProxy .
type APIProxy struct {
	apiDef    apiDefinition
	transport *transport
	balancer  Balancer
}

// MonitoringPath .
type MonitoringPath struct {
	UpstreamPath string
	Count        int64
	Duration     time.Duration
	AverageTime  int64
}

var globalMap = make(map[string]MonitoringPath)

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + "-" + t.apiDef.path + req.URL.Path
	start := time.Now()
	//response, err := http.DefaultTransport.RoundTrip(req)
	req.URL.Path = t.apiDef.upstreamPath + req.URL.Path
	response, err := t.tr.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	elapsed := time.Since(start)

	if val, ok := globalMap[key]; ok {
		val.Count = val.Count + 1
		val.Duration += time.Duration(elapsed.Nanoseconds())
		val.AverageTime = val.Duration.Nanoseconds() / val.Count
		globalMap[key] = val
		//do something here
	} else {
		var m MonitoringPath
		m.UpstreamPath = req.URL.Path
		m.Count = 1
		m.Duration = time.Duration(elapsed.Nanoseconds())
		m.AverageTime = val.Duration.Nanoseconds() / m.Count
		globalMap[key] = m
	}
	jsonMap, err := json.MarshalIndent(globalMap, "", "  ")
	if err == nil {
		logger.Debugf("Monitoring Graph: %s\n", jsonMap)
	}

	body, err := httputil.DumpResponse(response, true)
	if err != nil {
		logger.Error("error in dumb response")
		// copying the response body did not work
		return nil, err
	}

	logger.Debugf("Response Body: %+v", string(body))
	logger.Infof("Response Time: %s", time.Duration(elapsed.Nanoseconds()))
	return response, nil
}

func loadBalance(network, serviceName, serviceVersion string, apiDef apiDefinition) (net.Conn, error) {
	endpoints := apiDef.endpoints
	balancer := BalancerFor(apiDef.balancing)
	maxRetries := len(endpoints) * apiDef.maxRetries
	logger.Infof("max retries: %d * len(endpoints) = %d", apiDef.maxRetries, maxRetries)
	for retry := 1; retry <= maxRetries; retry++ {
		// selects the endpoint
		_, endpoint := balancer.Balance(endpoints)
		logger.Infof("balancing request to %s", endpoint)

		// try to connect
		conn, err := (&net.Dialer{
			Timeout:   time.Duration(config.Config.Gateway.Dial.UpstreamTimeout) * time.Second,
			KeepAlive: time.Duration(config.Config.Gateway.Dial.KeepAlive) * time.Second,
			DualStack: config.Config.Gateway.Dial.DualStack,
		}).Dial(network, endpoint)
		if err != nil {
			logger.Errorf("Error accessing %s/%s (%s): %s", serviceName, serviceVersion, endpoint, err)
			// retry connection to a different endpoint (according to the load balancing strategy)
			continue
		}
		// Success: return the connection.
		return conn, nil
	}
	// No available endpoint.
	return nil, fmt.Errorf("No endpoint available for %s/%s", serviceName, serviceVersion)
}

// Handler .
func (p *APIProxy) Handler(w http.ResponseWriter, req *http.Request) {
	logger.Infof("Proxy for %s to targets %+v with LB strategy %s", p.apiDef.path, p.apiDef.endpoints, p.apiDef.balancing)
	(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = p.apiDef.path
		},
		Transport: p.transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Errorf("gateway error: %s", err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(fmt.Sprintf("Bad Gateway - %s\n", err)))
		},
	}).ServeHTTP(w, req)
}

/*
maxIdleConnsPerHost: 1024
      idleConnTimeout: 90
      tlsHandshakeTimeout: 10
      expectContinueTimeout: 1*/
func newTransport(apiDef apiDefinition) *transport {
	transportConf := config.Config.Gateway.Transport
	tr := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			logger.Debug("dialing to upstream backend api service")
			addr = strings.Split(addr, ":")[0]
			tmp := strings.Split(addr, "/")
			if len(tmp) != 3 {
				return nil, ErrInvalidService
			}

			return LoadBalance(network, tmp[0], tmp[1], apiDef)
		},
		DisableKeepAlives:     transportConf.DisableKeepAlives,
		DisableCompression:    transportConf.DisableCompression,
		ResponseHeaderTimeout: time.Duration(transportConf.ResponseHeaderTimeout) * time.Second,
		MaxIdleConns:          transportConf.MaxIdleConnections,
		MaxIdleConnsPerHost:   transportConf.MaxIdleConnsPerHost,
		IdleConnTimeout:       time.Duration(transportConf.MaxIdleConnsPerHost) * time.Second,
		TLSHandshakeTimeout:   time.Duration(transportConf.TLSHandshakeTimeout) * time.Second,
		ExpectContinueTimeout: time.Duration(transportConf.ExpectContinueTimeout) * time.Second,
	}
	return &transport{tr: tr, apiDef: apiDef}
}

// NewProxy returns a new Proxy instance.
func NewProxy(apiDef apiDefinition) *APIProxy {
	return &APIProxy{
		apiDef: apiDef,
		//transport: &transport{http.DefaultTransport},
		transport: newTransport(apiDef),
	}
}
