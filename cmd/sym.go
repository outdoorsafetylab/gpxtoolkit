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
	symSymbol = ""
	symRegexp = ""
)

// symCmd represents the sym command
var symCmd = &cobra.Command{
	Use:   "sym",
	Short: "Alter the <sym> of GPX waypoints",
	Long:  `Alter the <sym> of GPX waypoints.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if symSymbol == "" {
			return fmt.Errorf("please specify the new symbol")
		}
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		if symRegexp != "" {
			re, err := regexp.Compile(symRegexp)
			if err != nil {
				return err
			}
			for _, wpt := range trackLog.WayPoints {
				if re.MatchString(wpt.GetName()) {
					wpt.Symbol = proto.String(symSymbol)
				}
			}
		} else {
			for _, wpt := range trackLog.WayPoints {
				wpt.Symbol = proto.String(symSymbol)
			}
		}
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(symCmd)
	symCmd.Flags().StringVarP(&symSymbol, "symbol", "s", symSymbol, "New symbol for the waypoints.")
	symCmd.Flags().StringVarP(&symRegexp, "regexp", "r", symRegexp, "Only alter waypoints which names match the given regular exporession.")
}
