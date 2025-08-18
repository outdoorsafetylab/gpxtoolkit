/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"
	"math"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	statsAlpha            = 0.2
	statsByTracks         = false
	statsCorrectElevation = false
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Calculate GPX statistics",
	Long:  `Calculate GPX statistics`,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		if statsCorrectElevation {
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
		}
		print := func(title string, st *gpx.TrackStats) {
			if title != "" {
				fmt.Fprintf(os.Stdout, "=== %s ===\n", title)
			}
			fmt.Fprintf(os.Stdout, "Start Time:     %v\n", st.StartTime().In(time.Local))
			fmt.Fprintf(os.Stdout, "Duration:       %v\n", st.Duration())
			fmt.Fprintf(os.Stdout, "Distance:       %v meter\n", math.Round(st.GetDistance()))
			fmt.Fprintf(os.Stdout, "Elevation/Min:  %v meter\n", math.Round(st.GetElevationMin()))
			fmt.Fprintf(os.Stdout, "Elevation/Max:  %v meter\n", math.Round(st.GetElevationMax()))
			fmt.Fprintf(os.Stdout, "Elevation/Avg:  %v meter\n", math.Round(st.GetElevationDistance()/st.GetDistance()))
			fmt.Fprintf(os.Stdout, "Elevation/Gain: %v meter\n", math.Round(st.GetElevationGain()))
			fmt.Fprintf(os.Stdout, "Elevation/Loss: %v meter\n", math.Round(st.GetElevationLoss()))
		}
		if statsByTracks {
			for i, t := range trackLog.Tracks {
				st, err := t.Stat(statsAlpha)
				if err != nil {
					return err
				}
				print(fmt.Sprintf("Track %d: %s", i, t.GetName()), st)
			}
		} else {
			st, err := trackLog.Stat(statsAlpha)
			if err != nil {
				return err
			}
			print("", st)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().Float64VarP(&statsAlpha, "alpha", "a", statsAlpha, "Alpha filter value for accumulating elevation gain and loss")
	statsCmd.Flags().BoolVarP(&statsByTracks, "tracks", "t", statsByTracks, "Calculate track by track")
	statsCmd.Flags().BoolVarP(&statsCorrectElevation, "correct-elevation", "e", statsCorrectElevation, "Correct elevation before calculation")
}
