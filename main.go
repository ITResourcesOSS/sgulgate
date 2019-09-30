package main

import "github.com/itross/sgulgate/cmd"

// import (
// 	"github.com/ITResourcesOSS/sgulgate/internal/config"
// 	"github.com/ITResourcesOSS/sgulgate/internal/gateway"
// )

// func main() {
// 	config.LoadConfiguration()
// 	gw := gateway.New()
// 	gw.PrintParams()
// 	gw.PrintApis()
// 	gw.Start()
// }

func main() {
	cmd.Execute()
}
