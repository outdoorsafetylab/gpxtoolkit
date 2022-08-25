/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"gpxtoolkit/gpx"

	"github.com/spf13/cobra"
)

// flattenCmd represents the flatten command
var flattenCmd = &cobra.Command{
	Use:   "flatten",
	Short: "Flatten GPX to a single track with a single segment",
	Long:  `Flatten GPX to a single track with a single segment`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		segment := &gpx.Segment{Points: make([]*gpx.Point, 0)}
		track := &gpx.Track{
			Segments: []*gpx.Segment{segment},
		}
		for _, t := range trackLog.Tracks {
			for _, s := range t.Segments {
				segment.Points = append(segment.Points, s.Points...)
			}
		}
		trackLog.Tracks = []*gpx.Track{track}
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(flattenCmd)
}
