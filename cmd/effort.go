/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"
	"os"

	"github.com/spf13/cobra"
)

var (
	effortAlpha                  = 0.2
	effortByTracks               = false
	effortCorrectElevation       = true
	effortWeightElevationGain    = 10.0
	effortWeightElevationLoss    = 3.3
	effortOxygenDensity          = true
	effortOxygenDensitySlope     = 0.0001621796175
	effortOxygenDensityIntercept = 0.9763291581
)

// effortCmd represents the effort command
var effortCmd = &cobra.Command{
	Use:   "effort",
	Short: "Calculate the kilimeter effort (KmE)",
	Long:  `Calculate the kilimeter effort (KmE)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		if effortCorrectElevation {
			elev := &gpxutil.CorrectElevation{
				Waypoints: false,
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
			dist := st.GetDistance() / 1000
			gain := st.GetElevationGain() / 1000
			loss := st.GetElevationLoss() / 1000
			effort := dist + gain*effortWeightElevationGain + loss*effortWeightElevationLoss
			if effortOxygenDensity {
				avg := st.GetElevationDistance() / st.GetDistance()
				effort *= effortOxygenDensitySlope*avg + effortOxygenDensityIntercept
				fmt.Fprintf(os.Stdout, "Formula: (dist + (gain × %.2f) + (loss × %.2f)) × (slope × avg_alt + intercept) = KmE\n", effortWeightElevationGain, effortWeightElevationLoss)
				fmt.Fprintf(os.Stdout, "         (%.2f + (%.2f × %.2f) + (%.2f × %.2f)) x (%f × %.2f + %f) = %.2f\n", dist, gain, effortWeightElevationGain, loss, effortWeightElevationLoss, effortOxygenDensitySlope, avg, effortOxygenDensityIntercept, effort)
			} else {
				fmt.Fprintf(os.Stdout, "Formula: dist + (gain × %.2f) + (loss × %.2f) = KmE\n", effortWeightElevationGain, effortWeightElevationLoss)
				fmt.Fprintf(os.Stdout, "         %.2f + (%.2f × %.2f) + (%.2f × %.2f) = %.2f\n", dist, gain, effortWeightElevationGain, loss, effortWeightElevationLoss, effort)
			}
		}
		if effortByTracks {
			for i, t := range trackLog.Tracks {
				st, err := t.Stat(effortAlpha)
				if err != nil {
					return err
				}
				print(fmt.Sprintf("Track %d: %s", i, t.GetName()), st)
			}
		} else {
			st, err := trackLog.Stat(effortAlpha)
			if err != nil {
				return err
			}
			print("", st)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(effortCmd)
	effortCmd.Flags().Float64VarP(&effortAlpha, "alpha", "a", effortAlpha, "Alpha filter value for accumulating elevation gain and loss")
	effortCmd.Flags().BoolVarP(&effortByTracks, "tracks", "t", effortByTracks, "Calculate track by track")
	effortCmd.Flags().BoolVarP(&effortCorrectElevation, "correct-elevation", "e", effortCorrectElevation, "Correct elevation before calculation")
	effortCmd.Flags().BoolVarP(&effortOxygenDensity, "oxygen-density", "o", effortOxygenDensity, "Consider oxygen density in the calculation")
}
