package registry

// Endpoint defines a service endpoint to proxy request to.
type Endpoint struct {
	// http protocol scheme: "http" or "https".
	Scheme string

	// ip address or hostname of the service endpoint.
	Host string

	// port of the service endpoint.
	Port int16
}

// ServiceInstance defines an instance of the service, and will be used as a
// reverse proxy endpoint for service requests.
type ServiceInstance struct {
	// Instance is the instance id of the service as registered in the discovery server.
	// It should be something like "service-name@service-host-name:service-port".
	InstanceID string

	// host and port to proxy requests to.
	Endpoint Endpoint
}

// ServiceDefinition defines the structure which identify a service in sgulgate.
type ServiceDefinition struct {
	// Name is the system global identifier for the service. Normally it should be
	// in the following form "service.name" from the microservice app configuration.
	Name string

	// RoutingID is the service id in incoming requests. It should be something similar
	// to the Name field value.
	// a.e. Name = "myservice-myServiceGroup" -> RoutingID = "myservice".
	// If RoutingID is empty, the proxy will use the Name field.
	RoutingID string

	// Service API version to proxy requests to. It will be composed with the APIPrefix.
	APIVersion string

	// Prefix for service API routes. Normally it will be "/api". It can be empty.
	APIPrefix string

	// Instnces is the pool of service instances actually up and running, to proxy requests to.
	Instances []ServiceInstance
}

// ServiceRegistry defines the contract for a struct to bee a Service Registry for the API Gateway.
type ServiceRegistry interface {
	Register()
}

// DefaultRegistry is the registry implementation for the gateway reverse proxy.
type DefaultRegistry struct {
	serviceDefinitions map[string]ServiceDefinition
}
