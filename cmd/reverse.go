/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"slices"

	"github.com/spf13/cobra"
)

var (
	reverseWaypoints = false
)

// reverseCmd represents the reverse command
var reverseCmd = &cobra.Command{
	Use:   "reverse",
	Short: "Reverse GPX tracks and waypoints",
	Long:  `Reverse GPX tracks and waypoints`,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		if reverseWaypoints {
			slices.Reverse(trackLog.WayPoints)
		}
		slices.Reverse(trackLog.Tracks)
		for _, t := range trackLog.Tracks {
			slices.Reverse(t.Segments)
			for _, s := range t.Segments {
				slices.Reverse(s.Points)
			}
		}
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(reverseCmd)
	reverseCmd.Flags().BoolVarP(&reverseWaypoints, "wpt", "w", reverseWaypoints, "Reverse waypoints")
}
