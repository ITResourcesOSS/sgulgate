package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ITResourcesOSS/sgul"
	"github.com/ITResourcesOSS/sgulgate/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// var conf *sgul.Configuration
var logger *zap.SugaredLogger

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "sgulgate",
	Short: "SGUL API Gateway",
	Long: `---------------------------------------------
sgul API Gateway v. 0.1.0
---------------------------------------------`,
}

// Execute .
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initialize)
}

func initialize() {
	config.LoadConfiguration()
	//logger = sgul.GetLoggerByConf(config.Config.Log).Sugar()
	logger = sgul.GetLogger().Sugar()

	env := strings.ToLower(os.Getenv("ENV"))
	if env != "dev" && env != "development" && env != "" {
		configureGentleShutdown()
	}
}

func configureGentleShutdown() {
	gracefulStop := make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		sig := <-gracefulStop
		logger.Infof("[%+v] SIGNAL CAUGHT", sig)

		logger.Info("wait for 2 second to finish processing")
		time.Sleep(2 * time.Second)

		logger.Info("service goes down")
		logger.Info("Bye!")
		logger.Sync()
		os.Exit(0)
	}()

	logger.Info("service gentle shutdown hook activated")
}
