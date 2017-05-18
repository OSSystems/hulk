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
	"github.com/OSSystems/pkg/log"
	"github.com/OSSystems/hulk/mqtt"
	"github.com/OSSystems/hulk/pkg/filewatcher"
	"github.com/Sirupsen/logrus"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	servicesDir   = "/etc/hulk.d/"
	brokerAddress = "tcp://localhost:1883"
	listenAddress = "unix:///var/run/hulkd.sock"
	authFile      = ""
	logLevel      = "info"
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
			log.SetLevel(logrus.InfoLevel)
		}

		client := newMqttClient()

		connectToBroker(client)

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

		watchAuthFile(hulk)

		hulk.Run()
	},
}

func main() {
	RootCmd.PersistentFlags().StringVarP(&servicesDir, "dir", "d", servicesDir, "Path to directory with services")
	RootCmd.PersistentFlags().StringVarP(&brokerAddress, "broker", "b", brokerAddress, "Broker address to connect")
	RootCmd.PersistentFlags().StringVarP(&listenAddress, "listen", "l", listenAddress, "API server listen address")
	RootCmd.PersistentFlags().StringVarP(&authFile, "auth", "a", authFile, "Authentication file")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", logLevel, "Set the logging level (panic|fatal|error|warn|info|debug)")

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newMqttClient() mqtt.MqttClient {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(brokerAddress)

	if _, err := os.Stat(authFile); err == nil {
		auth, err := godotenv.Read(authFile)
		if err == nil {
			log.WithFields(logrus.Fields{
				"file": authFile,
				"auth": auth,
			}).Debug("new authorization")

			if id, ok := auth["HULK_ID"]; ok {
				opts.SetClientID(id)
			}

			if username, ok := auth["HULK_USERNAME"]; ok {
				opts.SetUsername(username)
			}

			if password, ok := auth["HULK_PASSWORD"]; ok {
				opts.SetPassword(password)
			}
		}
	}

	return mqtt.NewPahoClient(opts)
}

func connectToBroker(client mqtt.MqttClient) {
	if err := client.Connect(); err != nil {
		log.WithFields(logrus.Fields{"reason": err}).Warn("Failed to connect to broker")
	}
}

func watchAuthFile(hulk *hulk.Hulk) {
	authWatcher, err := filewatcher.NewFileWatcher()
	if err != nil {
		log.Fatal(err)
	}

	if authFile != "" {
		if _, err := os.Stat(authFile); os.IsNotExist(err) {
			log.WithFields(logrus.Fields{"file": authFile}).Warn("auth file does not exist")
		}

		if err := authWatcher.Add(authFile); err != nil {
			log.Fatal(err)
		}
	}

	// Watches auth file for changes and reconnect to broker using the new credentials
	go authWatcher.Watch()
	go func() {
		for {
			select {
			case <-authWatcher.Changed:
				log.WithFields(logrus.Fields{"file": authFile}).Debug("auth file changed")

				client := newMqttClient()
				connectToBroker(client)

				if err = hulk.Reload(client); err != nil {
					log.Fatal(err)
				}
			}
		}
	}()
}
