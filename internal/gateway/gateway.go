package gateway

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/ITResourcesOSS/sgulgate/internal/config"
)

// ErrNoAPIFound returned if no API definition has been provvisioned for the request path.
var ErrNoAPIFound = errors.New("No API definition for request path")

type apiDefinition struct {
	name      string
	path      string
	balancing string
	endpoints []config.Target
}

// Gateway .
type Gateway struct {
	api     map[string]apiDefinition
	proxies map[string]*APIProxy
}

// New returns a new instance of the Gateway struct.
func New() Gateway {
	gw := Gateway{
		api:     make(map[string]apiDefinition),
		proxies: make(map[string]*APIProxy),
	}

	apiConf := config.Config.API
	log.Printf("configuring %s definitions", apiConf.Name)

	for _, endpoint := range apiConf.Endpoints {
		path := fmt.Sprintf("%s/v%s", endpoint.Path, endpoint.Version)
		apiDef := apiDefinition{
			name:      endpoint.Name,
			path:      path,
			balancing: endpoint.Proxy.Balancing.Strategy,
			endpoints: make([]config.Target, 0),
		}

		apiDef.endpoints = endpoint.Proxy.Targets
		// for _, target := range endpoint.Proxy.Targets {
		// 	apiDef.endpoints = append(
		// 		apiDef.endpoints,
		// 		fmt.Sprintf("%s%s", target.Host, target.Path))
		// 	// apiDef.endpoints = append(apiDef.endpoints, target.Host)
		// }
		gw.api[path] = apiDef
		gw.proxies[path] = NewProxy(apiDef)

		log.Printf("endpoint name: %s - path: %s - targets: %+v", apiDef.name, apiDef.path, apiDef.endpoints)
	}

	return gw
}

// PrintConfiguration .
func (gw Gateway) PrintConfiguration() {
	log.Printf("Gateway Configuation: %+v\n", config.Config)
}

// Start starts the Gateway starting the http server on configured endppoint.
func (gw Gateway) Start() {
	log.Println("starting Gateway...")
	log.Printf("gateway endpoint: %s", gw.endpointPath())

	http.HandleFunc(gw.endpointPath(), func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = gw.stripPath(req.URL.Path)
		name, version, err := gw.GetNameAndVersion(req.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		apiPath := fmt.Sprintf("/%s/%s", name, version)
		log.Printf("serving %s", apiPath)

		upstreamProxy := gw.proxies[apiPath]
		if upstreamProxy == nil {
			http.Error(w, ErrNoAPIFound.Error(), http.StatusNotFound)
			log.Printf("error serving request: %s", ErrNoAPIFound.Error())
			return
		}

		//upstreamProxy.Handler.ServeHTTP(w, req)
		upstreamProxy.Handler(w, req)
	})
	gw.serve()
}

func (gw Gateway) serve() {
	log.Printf("endpoint started and listening on localhost:9000%s", config.Config.Gateway.Endpoint.Path)
	log.Fatal(http.ListenAndServe(":9000", nil))
}

// GetNameAndVersion .
func (gw Gateway) GetNameAndVersion(target *url.URL) (name, version string, err error) {
	path := target.Path
	if len(path) > 1 && path[0] == '/' {
		path = path[1:]
	}
	tmp := strings.Split(path, "/")
	if len(tmp) < 2 {
		return "", "", fmt.Errorf("Invalid path")
	}
	name, version = tmp[0], tmp[1]
	target.Path = "/" + strings.Join(tmp[2:], "/")
	return name, version, nil
}

func sanitizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func (gw Gateway) endpointPath() string {
	epath := sanitizePath(config.Config.Gateway.Endpoint.Path)
	return fmt.Sprintf("/%s/", epath)
}

func (gw Gateway) stripPath(path string) string {
	epath := sanitizePath(config.Config.Gateway.Endpoint.Path)
	return strings.Replace(path, fmt.Sprintf("/%s", epath), "", -1)
}
