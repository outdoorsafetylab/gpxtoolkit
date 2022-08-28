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
	outlierSigma            = 6
	outlierDeduplicate      = true
	outlierCorrectElevation = true
	outlierByDistance       = false
	outlierByEIF            = 0.0
)

// outlierCmd represents the outlier command
var outlierCmd = &cobra.Command{
	Use:   "outlier",
	Short: "Remove outliers in GPX by sigma (standard deviation)",
	Long:  `Remove outliers in GPX by sigma (standard deviation)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		if outlierDeduplicate {
			dedup := gpxutil.RemoveDuplicated()
			n, err := dedup.Run(trackLog)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Removed %d duplications\n", n)
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
		if outlierByEIF > 0 {
			outlier := &gpxutil.RemoveOutlierByEIF{
				Threshold: outlierByEIF,
			}
			n, err := outlier.Run(trackLog)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Removed %d outliers by EIF\n", n)
		} else {
			var outlier *gpxutil.RemoveOutlier
			if outlierByDistance {
				outlier = gpxutil.RemoveOutlierByDistance(outlierSigma)
			} else {
				outlier = gpxutil.RemoveOutlierBySpeed(outlierSigma)
			}
			n, err := outlier.Run(trackLog)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Removed %d outliers\n", n)
		}
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(outlierCmd)
	outlierCmd.Flags().IntVarP(&outlierSigma, "sigma", "s", outlierSigma, "Level of sigma")
	outlierCmd.Flags().BoolVarP(&outlierDeduplicate, "deduplication", "d", outlierDeduplicate, "Remove duplicated points before calculation")
	outlierCmd.Flags().BoolVarP(&outlierCorrectElevation, "correct-elevation", "e", outlierCorrectElevation, "Correct elevation before calculation")
	outlierCmd.Flags().BoolVarP(&outlierByDistance, "distance", "D", outlierByDistance, "Calculate by distance instead of speed")
	outlierCmd.Flags().Float64Var(&outlierByEIF, "eif", outlierByEIF, "Remove outlier by EIF")
}
