/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"gpxtoolkit/gpxutil"

	"github.com/spf13/cobra"
)

var (
	elevIncludeWaypoints = false
)

// elevCmd represents the elev command
var elevCmd = &cobra.Command{
	Use:   "elev",
	Short: "Add or alter the elevation of GPX track points",
	Long:  `Add or alter the elevation of GPX track points.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		elev := &gpxutil.CorrectElevation{
			Waypoints: elevIncludeWaypoints,
			Service:   getElevationService(),
		}
		if elev.Service == nil {
			return fmt.Errorf("no elevation service")
		}
		_, err = elev.Run(trackLog)
		if err != nil {
			return err
		}
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(elevCmd)
	elevCmd.Flags().BoolVarP(&elevIncludeWaypoints, "waypoints", "w", elevIncludeWaypoints, "Include waypoints")
}
