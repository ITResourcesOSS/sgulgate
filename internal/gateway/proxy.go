package gateway

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrInvalidService returned if there is no api definition for service/version.
var ErrInvalidService = errors.New("invalid service/version")

// Proxy .
type Proxy struct {
	Handler http.HandlerFunc
}

// func loadBalance(network, serviceName, serviceVersion string, reg registry.Registry) (net.Conn, error) {
// 	endpoints, err := reg.Lookup(serviceName, serviceVersion)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for {
// 		// No more endpoint, stop
// 		if len(endpoints) == 0 {
// 			break
// 		}
// 		// Select a random endpoint
// 		i := rand.Int() % len(endpoints)
// 		endpoint := endpoints[i]

// 		// Try to connect
// 		conn, err := net.Dial(network, endpoint)
// 		if err != nil {
// 			reg.Failure(serviceName, serviceVersion, endpoint, err)
// 			// Failure: remove the endpoint from the current list and try again.
// 			endpoints = append(endpoints[:i], endpoints[i+1:]...)
// 			continue
// 		}
// 		// Success: return the connection.
// 		return conn, nil
// 	}
// 	// No available endpoint.
// 	return nil, fmt.Errorf("No endpoint available for %s/%s", serviceName, serviceVersion)
// }

// NewProxy returns a new Proxy instance.
func NewProxy(apiDef apiDefinition) *Proxy {
	// transport := &http.Transport{
	// 	Proxy: http.ProxyFromEnvironment,
	// 	Dial: func(network, addr string) (net.Conn, error) {
	// 		addr = strings.Split(addr, ":")[0]
	// 		tmp := strings.Split(addr, "/")
	// 		if len(tmp) != 2 {
	// 			return nil, ErrInvalidService
	// 		}
	// 		return LoadBalance(network, tmp[0], tmp[1], reg)
	// 	},
	// 	TLSHandshakeTimeout: 10 * time.Second,
	// }
	return &Proxy{
		Handler: func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Proxy for %s to targets %+v with LB strategy %s", apiDef.path, apiDef.endpoints, apiDef.balancing)))
		},
	}
}
