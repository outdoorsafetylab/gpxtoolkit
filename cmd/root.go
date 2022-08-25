/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/log"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	file           string
	elevationURL   string
	elevationToken string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gpxtoolkit",
	Short: "A swiss knife for processing GPX files",
	Long:  `A swiss knife for processing GPX files`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func loadGpx() (*gpx.TrackLog, error) {
	var f io.ReadCloser
	var err error
	parser := &gpx.Parser{}
	if file != "" {
		f, err = os.Open(file)
		if err != nil {
			return nil, err
		}
	} else {
		f = os.Stdin
	}
	gpxLog, err := parser.Parse(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open GPX '%s': %s\n", file, err.Error())
		return nil, err
	}
	return gpxLog, nil
}

func dumpGpx(gpxLog *gpx.TrackLog) error {
	writer := &gpx.Writer{
		Creator: rootCmd.Use,
		Writer:  os.Stdout,
	}
	err := writer.Write(gpxLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write GPX: %s\n", err.Error())
		return err
	}
	return nil
}

func getElevationService() elevation.Service {
	if elevationURL == "" {
		elevationURL = os.Getenv("ELEVATION_URL")
		if elevationURL != "" {
			log.Infof("Using elevation service URL from environment variable: %s", elevationURL)
		}
	} else if elevationURL != "" {
		log.Infof("Using elevation service URL: %s", elevationURL)
	}
	if elevationToken == "" {
		elevationToken = os.Getenv("ELEVATION_TOKEN")
		if elevationToken != "" {
			log.Infof("Using elevation service token from environment variable")
		}
	}
	if elevationURL != "" {
		return &elevation.OutdoorSafetyLab{
			Client: http.DefaultClient,
			URL:    elevationURL,
			Token:  elevationToken,
		}
	} else {
		log.Infof("No elevation service")
		return nil
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "GPX file name; will read from stdin if this is not specified")
	rootCmd.PersistentFlags().StringVar(&elevationURL, "elevation-url", "", "URL for elevation service")
	rootCmd.PersistentFlags().StringVar(&elevationToken, "elevation-token", "", "auth token of elevation service")
}
