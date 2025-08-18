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
	timeshiftDay      = 0
	timeshiftDuration = "0s"
)

// timeshiftCmd represents the time command
var timeshiftCmd = &cobra.Command{
	Use:   "timeshift",
	Args:  cobra.NoArgs,
	Short: "Shift the time of GPX way points and track points",
	Long:  `Shift the time of GPX way points and track points.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		duration, err := time.ParseDuration(timeshiftDuration)
		if err != nil {
			return err
		}
		duration += time.Hour * 24 * time.Duration(timeshiftDay)
		time := &gpxutil.TimeShift{
			Duration: duration,
		}
		_, err = time.Run(trackLog)
		if err != nil {
			return err
		}
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(timeshiftCmd)
	timeshiftCmd.Flags().StringVarP(&timeshiftDuration, "duration", "d", timeshiftDuration, "Add specified duration to all points in Golang's duration string (can be nagetive)")
	timeshiftCmd.Flags().IntVarP(&timeshiftDay, "day", "D", timeshiftDay, "Add specified duration in day to all points (can be nagetive)")
}
