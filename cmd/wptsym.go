/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var (
	wptsymSymbol = ""
	wptsymRegexp = ""
)

// wptsymCmd represents the wptsym command
var wptsymCmd = &cobra.Command{
	Use:   "wptsym",
	Short: "Alter the <sym> of GPX waypoints",
	Long:  `Alter the <sym> of GPX waypoints.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if wptsymSymbol == "" {
			return fmt.Errorf("please specify the new symbol")
		}
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		if wptsymRegexp != "" {
			re, err := regexp.Compile(wptsymRegexp)
			if err != nil {
				return err
			}
			for _, wpt := range trackLog.WayPoints {
				if re.MatchString(wpt.GetName()) {
					wpt.Symbol = proto.String(wptsymSymbol)
				}
			}
		} else {
			for _, wpt := range trackLog.WayPoints {
				wpt.Symbol = proto.String(wptsymSymbol)
			}
		}
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(wptsymCmd)
	wptsymCmd.Flags().StringVarP(&wptsymSymbol, "symbol", "s", wptsymSymbol, "New symbol for the waypoints.")
	wptsymCmd.Flags().StringVarP(&wptsymRegexp, "regexp", "r", wptsymRegexp, "Only alter waypoints which names match the given regular exporession.")
}
