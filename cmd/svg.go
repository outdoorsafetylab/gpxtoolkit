/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"gpxtoolkit/cmd/svg"

	"github.com/spf13/cobra"
)

var createSVG = &svg.Create{
	ZoomLevel:  16,
	TileWidth:  256,
	TileHeight: 256,
	Background: "#a0ffa0",
	TilePadding: struct {
		Top     int
		Left    int
		Botttom int
		Right   int
	}{
		Top:     0,
		Left:    0,
		Botttom: 0,
		Right:   0,
	},
	Scale: struct {
		Unit       int
		Repeat     int
		Stroke     string
		FillColors []string
	}{
		Unit:       500,
		Repeat:     5,
		Stroke:     "#000000",
		FillColors: []string{"#a0a0a0", "#ffffff"},
	},
}

// svgCmd represents the svg command
var svgCmd = &cobra.Command{
	Use:   "svg",
	Short: "Create a SVG image with the GPX and the map of its coverage.",
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		return createSVG.Run(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(svgCmd)
	svgCmd.Flags().IntVarP(&createSVG.ZoomLevel, "zoom", "z", createSVG.ZoomLevel, "Zoom level")
	svgCmd.Flags().StringVarP(&createSVG.Background, "background", "g", createSVG.Background, "Background color")
	svgCmd.Flags().BoolVarP(&createSVG.EmbedImage, "embed", "e", createSVG.EmbedImage, "Embed image into svg")
	svgCmd.Flags().IntVarP(&createSVG.TilePadding.Top, "top", "t", createSVG.TilePadding.Top, "Top padding, in tiles.")
	svgCmd.Flags().IntVarP(&createSVG.TilePadding.Left, "left", "l", createSVG.TilePadding.Left, "Left padding, in tiles.")
	svgCmd.Flags().IntVarP(&createSVG.TilePadding.Botttom, "bottom", "b", createSVG.TilePadding.Botttom, "Bottom padding, in tiles.")
	svgCmd.Flags().IntVarP(&createSVG.TilePadding.Right, "right", "r", createSVG.TilePadding.Right, "Right padding, in tiles.")
}
