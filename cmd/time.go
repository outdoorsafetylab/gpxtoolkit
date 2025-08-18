/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"gpxtoolkit/gpxutil"
	"time"

	"github.com/spf13/cobra"
)

var (
	timeStart       = "(now)"
	timeSpeed       = 10.0
	terrainDistance = false
)

// timeCmd represents the time command
var timeCmd = &cobra.Command{
	Use:   "time",
	Args:  cobra.NoArgs,
	Short: "Add or alter the time of GPX track points",
	Long:  `Add or alter the time of GPX track points.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		start := time.Now()
		if timeStart != "(now)" {
			start, err = time.Parse("2006-01-02 15:04:05", timeStart)
			if err != nil {
				return err
			}
		}
		time := &gpxutil.ReTimestamp{
			DistanceFunc: gpxutil.HaversinDistance,
			Start:        start,
			Speed:        timeSpeed,
		}
		if terrainDistance {
			time.DistanceFunc = gpxutil.TerrainDistance
		}
		_, err = time.Run(trackLog)
		if err != nil {
			return err
		}
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(timeCmd)
	timeCmd.Flags().StringVarP(&timeStart, "start", "S", timeStart, "Start time for deriving the time, in YYYY-MM-DD HH:mm:SS")
	timeCmd.Flags().Float64VarP(&timeSpeed, "speed", "s", timeSpeed, "Average speed for deriving the time")
	timeCmd.Flags().BoolVarP(&terrainDistance, "terrain", "t", terrainDistance, "Use terrain (3D) distance instead of haversin (2D) distance")
}
