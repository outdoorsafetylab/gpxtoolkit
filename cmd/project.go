/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"gpxtoolkit/gpxutil"
	"os"

	"github.com/spf13/cobra"
)

var (
	projectThreshold    float64 = 30
	projectKeepOriginal bool
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Project waypoints onto tracks",
	Long:  `Project waypoints onto tracks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		project := &gpxutil.ProjectWaypoints{
			DistanceFunc: gpxutil.HaversinDistance,
			Threshold:    projectThreshold,
			KeepOriginal: projectKeepOriginal,
		}
		n, err := project.Run(trackLog)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Projected %d waypoints\n", n)
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.Flags().BoolVarP(&projectKeepOriginal, "keep", "k", projectKeepOriginal, "Keep the original waypoints")
	projectCmd.Flags().Float64VarP(&projectThreshold, "threshold", "t", projectThreshold, "Distance threshold of waypoints. Waypoints farer than this threshold won't be used for projection.")
}
