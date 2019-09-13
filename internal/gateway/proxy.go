package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// ErrInvalidService returned if there is no api definition for service/version.
var ErrInvalidService = errors.New("invalid service/version")

// LoadBalance is the load balancing startegy implementation.
var LoadBalance = loadBalance

type transport struct {
	*http.Transport
}

// Proxy .
type Proxy interface {
	Handler(w http.ResponseWriter, req *http.Request)
}

// APIProxy .
type APIProxy struct {
	apiDef    apiDefinition
	transport *transport
}

// MonitoringPath .
type MonitoringPath struct {
	Path        string
	Count       int64
	Duration    time.Duration
	AverageTime int64
}

var globalMap = make(map[string]MonitoringPath)

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> RoundTrip")
	start := time.Now()
	response, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		print("\n\ncame in error resp here", err)
		return nil, err
	}
	elapsed := time.Since(start)

	key := req.Method + "-" + req.URL.Path

	if val, ok := globalMap[key]; ok {
		val.Count = val.Count + 1
		val.Duration += time.Duration(elapsed.Nanoseconds())
		val.AverageTime = val.Duration.Nanoseconds() / val.Count
		globalMap[key] = val
		//do something here
	} else {
		var m MonitoringPath
		m.Path = req.URL.Path
		m.Count = 1
		m.Duration = time.Duration(elapsed.Nanoseconds())
		m.AverageTime = val.Duration.Nanoseconds() / m.Count
		globalMap[key] = m
	}
	jsonMap, err := json.MarshalIndent(globalMap, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	log.Printf("Monitoring Graph: %s\n", jsonMap)

	body, err := httputil.DumpResponse(response, true)
	if err != nil {
		print("\n\nerror in dumb response")
		// copying the response body did not work
		return nil, err
	}

	log.Println("Response Body : ", string(body))
	log.Println("Response Time:", time.Duration(elapsed.Nanoseconds()))
	return response, nil
}

func loadBalance(network, serviceName, serviceVersion string, apiDef apiDefinition) (net.Conn, error) {
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> loadBalance")
	// endpoints, err := reg.Lookup(serviceName, serviceVersion)
	endpoints := apiDef.endpoints
	// if err != nil {
	// 	return nil, err
	// }
	for {
		// No more endpoint, stop
		if len(endpoints) == 0 {
			break
		}
		// Select a random endpoint
		i := rand.Intn(100) % len(endpoints)
		endpoint := endpoints[i]

		// Try to connect
		conn, err := net.Dial(network, endpoint.Host)
		if err != nil {
			// reg.Failure(serviceName, serviceVersion, endpoint, err)
			log.Printf("Error accessing %s/%s (%s): %s", serviceName, serviceVersion, endpoint, err)
			// Failure: remove the endpoint from the current list and try again.
			endpoints = append(endpoints[:i], endpoints[i+1:]...)
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
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> Handler")
	log.Printf("Proxy for %s to targets %+v with LB strategy %s", p.apiDef.path, p.apiDef.endpoints, p.apiDef.balancing)
	(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> Director")
			req.URL.Scheme = "http"
			req.URL.Host = p.apiDef.endpoints[0].Host
			req.URL.Path = p.apiDef.endpoints[0].Path + req.URL.Path
			req.Header.Add("X-Forwarded-Host", req.Host)
			origin, _ := url.Parse(p.apiDef.endpoints[0].Host)
			req.Header.Add("X-Origin-Host", origin.Host)
		},
		Transport: p.transport,
	}).ServeHTTP(w, req)
}

func newTransport(apiDef apiDefinition) *transport {
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> newTrasport")
	tr := &http.Transport{
		DisableKeepAlives:     true,
		MaxIdleConnsPerHost:   100000,
		DisableCompression:    true,
		ResponseHeaderTimeout: 30 * time.Second,
		Dial: func(network, addr string) (net.Conn, error) {
			log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> DIAL")
			addr = strings.Split(addr, ":")[0]
			tmp := strings.Split(addr, "/")
			if len(tmp) != 2 {
				return nil, ErrInvalidService
			}
			return LoadBalance(network, tmp[0], tmp[1], apiDef)
		},
		// Dial: (&net.Dialer{
		// 	Timeout:   30 * time.Second,
		// 	KeepAlive: 30 * time.Second,
		// }).Dial,
	}
	return &transport{tr}
}

// NewProxy returns a new Proxy instance.
func NewProxy(apiDef apiDefinition) *APIProxy {
	return &APIProxy{
		apiDef: apiDef,
		//transport: &transport{http.DefaultTransport},
		transport: newTransport(apiDef),
	}
}
