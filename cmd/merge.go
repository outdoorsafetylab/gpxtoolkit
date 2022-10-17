/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"gpxtoolkit/gpx"

	"github.com/spf13/cobra"
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge",
	Args:  cobra.NoArgs,
	Short: "Merge multiple GPX track logs as a single one",
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLogs, err := loadTrackLogs()
		if err != nil {
			return err
		}
		if len(trackLogs) < 2 {
			err := fmt.Errorf("At least 2 GPX files must be provided")
			return err
		}
		merged := &gpx.TrackLog{}
		for _, trackLog := range trackLogs {
			merged.WayPoints = append(merged.WayPoints, trackLog.WayPoints...)
			merged.Tracks = append(merged.Tracks, trackLog.Tracks...)
		}
		return dumpGpx(merged)
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}
