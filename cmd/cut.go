/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"

	"github.com/spf13/cobra"
)

var (
	cutThreshold float64 = 30
	cutWaypoints []string
)

// cutCmd represents the cut command
var cutCmd = &cobra.Command{
	Use:   "cut",
	Short: "Cut GPX by waypoints",
	Long:  `Cut GPX to tracks by waypoints`,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		cut := &gpxutil.SliceByWaypoints{
			DistanceFunc: gpxutil.HaversinDistance,
			Threshold:    cutThreshold,
		}
		if len(cutWaypoints) > 0 {
			cut.Waypoints = make([]*gpx.WayPoint, len(cutWaypoints))
			for i, w := range cutWaypoints {
				for _, ww := range trackLog.WayPoints {
					if w != ww.GetName() {
						continue
					}
					cut.Waypoints[i] = ww
					break
				}
				if cut.Waypoints[i] == nil {
					return fmt.Errorf("no such waypoint: %s", w)
				}
			}
		} else {
			cut.Waypoints = trackLog.WayPoints
		}
		_, err = cut.Run(trackLog)
		if err != nil {
			return err
		}
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(cutCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cutCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cutCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	cutCmd.Flags().StringSliceVarP(&cutWaypoints, "waypoint", "w", cutWaypoints, "Distance threshold of waypoints. Waypoints farer than this threshold won't be used for cutting.")
	cutCmd.Flags().Float64VarP(&cutThreshold, "threshold", "t", cutThreshold, "Distance threshold of waypoints. Waypoints farer than this threshold won't be used for cutting.")
}
