package main

import (
	"errors"
	"os"

	"github.com/OSSystems/hulk/client"
	"github.com/OSSystems/hulk/log"
	"github.com/spf13/cobra"
)

var (
	hulkAddress = "unix:///var/run/hulkd.sock"
)

var cli *client.Client

var rootCmd = &cobra.Command{
	Use:   "hulkctl [OPTIONS] COMMAND [arg...]",
	Short: "Hulk Control",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initClient()
	},
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&hulkAddress, "address", "a", hulkAddress, "Hulk Daemon host address")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "list-services",
		Short: "List Hulk services",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initClient()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ListServices(cli)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "inspect-service [NAME]",
		Short: "Inspect Hulk service",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initClient()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("Service name is missing")
			}

			return InspectService(cli, args[0])
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initClient() {
	if hulkEnvAddress := os.Getenv("HULK_ADDRESS"); hulkEnvAddress != "" {
		hulkAddress = hulkEnvAddress
	}

	var err error
	if cli, err = client.NewClient(hulkAddress); err != nil {
		log.Fatal(err)
	}
}
