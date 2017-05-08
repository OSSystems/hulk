package main

import (
	"fmt"
	"os"

	"github.com/OSSystems/hulk/hulk"
	"github.com/OSSystems/hulk/log"
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
		switch logLevel {
		case "panic":
			log.SetLevel(logrus.PanicLevel)
		case "fatal":
			log.SetLevel(logrus.FatalLevel)
		case "error":
			log.SetLevel(logrus.ErrorLevel)
		case "warn":
			log.SetLevel(logrus.WarnLevel)
		case "info":
			log.SetLevel(logrus.InfoLevel)
		case "debug":
			log.SetLevel(logrus.DebugLevel)
		default:
			log.SetLevel(logrus.WarnLevel)
		}

		opts := MQTT.NewClientOptions()
		opts.AddBroker(brokerAddress)

		client := mqtt.NewPahoClient(opts)

		if err := client.Connect(); err != nil {
			log.Fatal(err)
		}

		hulk, err := hulk.NewHulk(client, servicesDir)
		if err != nil {
			log.Fatal(err)
		}

		if err := hulk.LoadServices(); err != nil {
			log.Fatal(err)
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
