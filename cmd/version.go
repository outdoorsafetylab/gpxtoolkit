package cmd

import (
	"fmt"
	"gpxtoolkit/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gpxtoolkit",
	Long:  `Print the version number of gpxtoolkit`,
	Run: func(cmd *cobra.Command, args []string) {
		if version.GitTag != "" {
			fmt.Printf("gpxtoolkit %s (%s)\n", version.GitTag, version.GitHash)
		} else {
			fmt.Printf("gpxtoolkit %s\n", version.GitHash)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
