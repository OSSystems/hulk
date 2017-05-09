package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/OSSystems/hulk/api/server"
	"github.com/OSSystems/hulk/api/server/router"
	"github.com/OSSystems/hulk/api/server/router/service"
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
	listenAddress = "unix:///var/run/hulkd.sock"
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

		routes := []router.Route{}
		routes = append(routes, service.Routes(hulk)...)

		listener, err := server.NewListener(listenAddress)
		if err != nil {
			log.Fatal(err)
		}

		log.Infof("Hulk API listening on %s", listenAddress)

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)

		go func() {
			<-sigc
			listener.Close()
			os.Exit(0)
		}()

		go server.Listen(listener, router.NewRouter(routes))

		hulk.Run()
	},
}

func main() {
	RootCmd.PersistentFlags().StringVarP(&servicesDir, "dir", "d", servicesDir, "Path to directory with services")
	RootCmd.PersistentFlags().StringVarP(&brokerAddress, "broker", "b", brokerAddress, "Broker address to connect")
	RootCmd.PersistentFlags().StringVarP(&listenAddress, "listen", "l", listenAddress, "API server listen address")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", logLevel, "Set the logging level (panic|fatal|error|warn|info|debug)")

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
