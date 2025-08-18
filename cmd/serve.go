/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"strconv"

	"gpxtoolkit/log"
	"gpxtoolkit/server"

	"github.com/spf13/cobra"
)

var daemon = &server.Server{
	Webroot: "./webroot/dist",
	Port:    8080,
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run as a web application",
	Long:  `Run as a web application`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dev, err := cmd.Flags().GetBool("debug")
		if err != nil {
			return err
		}
		err = log.Init(dev)
		if err != nil {
			return err
		}
		env := os.Getenv("PORT")
		if env != "" {
			val, err := strconv.ParseInt(env, 10, 32)
			if err == nil {
				log.Infof("Using HTTP port from environment variable: %s", env)
				daemon.Port = int16(val)
			}
		}
		return daemon.Run(getElevationService())
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serveCmd.Flags().BoolP("debug", "d", false, "Run in development mode")
	serveCmd.Flags().Int16VarP(&daemon.Port, "port", "p", daemon.Port, "Port for service HTTP")
	serveCmd.Flags().StringVarP(&daemon.Webroot, "webroot", "w", daemon.Webroot, "Root folder for static web content")
}
