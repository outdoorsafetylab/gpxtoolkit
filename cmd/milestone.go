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
	milestoneDistance        = 100.0
	milestoneTemplate        = "printf(\"%.1fK\", dist/1000)"
	milestoneSymbol          = "Milestone"
	milestoneReverse         = false
	milestoneFits            = false
	milestoneTerrainDistance = false
	milestoneFormat          = "gpx"
)

// milestoneCmd represents the milestone command
var milestoneCmd = &cobra.Command{
	Use:   "milestone",
	Short: "Create milestone waypoints along GPX tracks",
	Long: `Create milestone waypoints along GPX tracks at specified distance intervals.

This command reads a GPX file and creates milestone waypoints at regular distance
intervals along the tracks. The milestones can be customized with templates for
naming and various options for placement.

Examples:
  # Create milestones every 100 meters
  gpxtoolkit milestone --file track.gpx --distance 100

  # Use custom template for milestone names
  gpxtoolkit milestone --file track.gpx --template 'printf("%.1fK", dist/1000)'

  # Create milestones with custom symbol
  gpxtoolkit milestone --file track.gpx --symbol "Flag" --distance 200

  # Use terrain distance instead of 2D distance
  gpxtoolkit milestone --file track.gpx --terrain-distance --distance 100

  # Output as CSV format
  gpxtoolkit milestone --file track.gpx --format csv
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load GPX file
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}

		// Validate template
		name := &gpxutil.MilestoneName{
			Template: milestoneTemplate,
		}
		_, err = name.Eval(&gpxutil.MilestoneNameVariables{})
		if err != nil {
			return fmt.Errorf("invalid template: %w", err)
		}

		// Create milestone command
		commands := &gpxutil.ChainedCommands{
			Commands: []gpxutil.Command{
				gpxutil.RemoveDistanceLessThan(0.1),
				&gpxutil.Milestone{
					Service:           getElevationService(),
					Distance:          milestoneDistance,
					MilestoneName:     name,
					Reverse:           milestoneReverse,
					Symbol:            milestoneSymbol,
					FitWaypoints:      milestoneFits,
					ByTerrainDistance: milestoneTerrainDistance,
				},
			},
		}

		// Execute milestone creation
		_, err = commands.Run(trackLog)
		if err != nil {
			return fmt.Errorf("failed to create milestones: %w", err)
		}

		// Output based on format
		switch milestoneFormat {
		case "gpx":
			return dumpGpx(trackLog)
		case "csv":
			csv := gpxutil.NewCSVWayPointWriter(os.Stdout)
			_, err = csv.Run(trackLog)
			if err != nil {
				return fmt.Errorf("failed to write CSV: %w", err)
			}
			return nil
		default:
			return fmt.Errorf("unknown format: %s", milestoneFormat)
		}
	},
}

func init() {
	rootCmd.AddCommand(milestoneCmd)

	// Local flags for milestone command
	milestoneCmd.Flags().Float64VarP(&milestoneDistance, "distance", "d", milestoneDistance, "Distance between milestones in meters")
	milestoneCmd.Flags().StringVarP(&milestoneTemplate, "template", "t", milestoneTemplate, "Template for milestone names (Go printf syntax)")
	milestoneCmd.Flags().StringVarP(&milestoneSymbol, "symbol", "s", milestoneSymbol, "Symbol for milestone waypoints")
	milestoneCmd.Flags().BoolVarP(&milestoneReverse, "reverse", "r", milestoneReverse, "Create milestones in reverse order")
	milestoneCmd.Flags().BoolVar(&milestoneFits, "fits", milestoneFits, "Fit milestones to existing waypoints")
	milestoneCmd.Flags().BoolVarP(&milestoneTerrainDistance, "terrain-distance", "e", milestoneTerrainDistance, "Use terrain distance instead of 2D distance")
	milestoneCmd.Flags().StringVarP(&milestoneFormat, "format", "o", milestoneFormat, "Output format (gpx or csv)")
}
