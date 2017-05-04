package main

import (
	"fmt"
	"os"

	"github.com/OSSystems/hulk/hulk"
	"github.com/OSSystems/hulk/mqtt"
	"github.com/Sirupsen/logrus"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
)

var (
	servicesDir   = "/etc/hulk.d/"
	brokerAddress = "tcp://localhost:1883"
	logLevel      = "warn"
)

var RootCmd = &cobra.Command{
	Use:   "hulkd",
	Short: "Hulk Daemon",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()

		switch logLevel {
		case "panic":
			logger.Level = logrus.PanicLevel
		case "fatal":
			logger.Level = logrus.FatalLevel
		case "error":
			logger.Level = logrus.ErrorLevel
		case "warn":
			logger.Level = logrus.WarnLevel
		case "info":
			logger.Level = logrus.InfoLevel
		case "debug":
			logger.Level = logrus.DebugLevel
		default:
			logger.Level = logrus.WarnLevel
		}

		opts := MQTT.NewClientOptions()
		opts.AddBroker(brokerAddress)

		client := mqtt.NewPahoClient(opts)

		if err := client.Connect(); err != nil {
			logger.Fatal(err)
		}

		hulk, err := hulk.NewHulk(client, servicesDir, logger)
		if err != nil {
			logger.Fatal(err)
		}

		if err := hulk.LoadServices(); err != nil {
			logger.Fatal(err)
		}

		hulk.Run()
	},
}

func main() {
	RootCmd.PersistentFlags().StringVarP(&servicesDir, "dir", "d", servicesDir, "Path to directory with services")
	RootCmd.PersistentFlags().StringVarP(&brokerAddress, "broker", "b", brokerAddress, "Broker address to connect")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", logLevel, "Set the logging level (panic|fatal|error|warn|info|debug)")

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
