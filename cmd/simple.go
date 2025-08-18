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
	simpleEpsilon = 10.0
	simpleFirst   = true
)

// simpleCmd represents the simple command
var simpleCmd = &cobra.Command{
	Use:   "simple",
	Short: "Simplify GPX points",
	Long:  `Simplify GPX points.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		simplify := &gpxutil.Simplify{
			Epsilon: simpleEpsilon,
			First:   simpleFirst,
			Service: getElevationService(),
		}
		n, err := simplify.Run(trackLog)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Simplified %d points\n", n)
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(simpleCmd)
	simpleCmd.Flags().BoolVarP(&simpleFirst, "first", "F", simpleFirst, "Simplify the first point")
	simpleCmd.Flags().Float64VarP(&simpleEpsilon, "epsilon", "e", simpleEpsilon, "Epsilon (distance) for simplification")
}
