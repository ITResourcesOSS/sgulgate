package main

import (
	"github.com/ITResourcesOSS/sgulgate/internal/config"
	"github.com/ITResourcesOSS/sgulgate/internal/gateway"
)

func main() {
	config.LoadConfiguration()
	gw := gateway.New()
	gw.Start()

}
