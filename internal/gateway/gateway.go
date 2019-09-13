package gateway

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/ITResourcesOSS/sgulgate/internal/config"
)

// Key to use when setting the request ID.
type ctxKeyRequestID int

// RequestIDKey is the key that holds the unique request ID in a request context.
const RequestIDKey ctxKeyRequestID = 0

// ErrNoAPIFound returned if no API definition has been provvisioned for the request path.
var ErrNoAPIFound = errors.New("No API definition for request path")

var requestIDPrefix string

type apiDefinition struct {
	name           string
	path           string
	balancing      string
	upstreamPath   string
	upstreamSchema string
	endpoints      []string
}

// Gateway .
type Gateway struct {
	api     map[string]apiDefinition
	proxies map[string]*APIProxy
}

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	requestIDPrefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
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
			name:           endpoint.Name,
			path:           path,
			balancing:      endpoint.Proxy.Balancing.Strategy,
			upstreamPath:   endpoint.Proxy.Path,
			upstreamSchema: endpoint.Proxy.Schema,
			endpoints:      make([]string, 0),
		}

		apiDef.endpoints = endpoint.Proxy.Targets
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

func requestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestUUID, _ := uuid.NewUUID()
			requestID = fmt.Sprintf("sgulgate@%s-%s", requestIDPrefix, requestUUID.String())
		}
		ctx = context.WithValue(ctx, RequestIDKey, requestID)
		r.Header.Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// Start starts the Gateway starting the http server on configured endppoint.
func (gw Gateway) Start() {
	log.Println("starting Gateway...")
	log.Printf("gateway endpoint: %s", gw.endpointPath())

	http.HandleFunc(gw.endpointPath(), requestIDMiddleware(func(w http.ResponseWriter, req *http.Request) {
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
	}))
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
