package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/outdoorsafetylab/gpxtoolkit/csvgpx"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var (
	csv2gpxDefaults      = csvgpx.DefaultOptions()
	csv2gpxDate          = csv2gpxDefaults.Date
	csv2gpxTime          = csv2gpxDefaults.Time
	csv2gpxTimeFormat    = csv2gpxDefaults.TimeFormat
	csv2gpxTimeZone      = csv2gpxDefaults.TimeZone
	csv2gpxLatitude      = csv2gpxDefaults.Latitude
	csv2gpxLongitude     = csv2gpxDefaults.Longitude
	csv2gpxElevation     = csv2gpxDefaults.Elevation
	csv2gpxName          = csv2gpxDefaults.Name
	csv2gpxNamePattern   = csv2gpxDefaults.NamePattern
	csv2gpxComment       = csv2gpxDefaults.Comment
	csv2gpxDescription   = csv2gpxDefaults.Description
	csv2gpxWaypointsOnly = csv2gpxDefaults.WaypointsOnly
	csv2gpxTWD97         = csv2gpxDefaults.TWD97
	csv2gpxEasting       = csv2gpxDefaults.Easting
	csv2gpxNorthing      = csv2gpxDefaults.Northing
)

// csv2gpxCmd represents the csv2gpx command
var csv2gpxCmd = &cobra.Command{
	Use:   "csv2gpx",
	Short: "Convert CSV to GPX format",
	RunE: func(cmd *cobra.Command, args []string) error {
		var reader io.Reader
		var name *string
		switch len(files) {
		case 0:
			reader = os.Stdin
		case 1:
			f, err := os.Open(files[0])
			if err != nil {
				return err
			}
			defer f.Close()
			name = proto.String(files[0])
			reader = f
		default:
			return fmt.Errorf("can only convert one CSV file for each time")
		}
		trackLog, err := csvgpx.Convert(reader, &csvgpx.Options{
			Date:          csv2gpxDate,
			Time:          csv2gpxTime,
			TimeFormat:    csv2gpxTimeFormat,
			TimeZone:      csv2gpxTimeZone,
			Latitude:      csv2gpxLatitude,
			Longitude:     csv2gpxLongitude,
			Elevation:     csv2gpxElevation,
			Name:          csv2gpxName,
			NamePattern:   csv2gpxNamePattern,
			Comment:       csv2gpxComment,
			Description:   csv2gpxDescription,
			WaypointsOnly: csv2gpxWaypointsOnly,
			TWD97:         csv2gpxTWD97,
			Easting:       csv2gpxEasting,
			Northing:      csv2gpxNorthing,
			Service:       getElevationService(),
		})
		if err != nil {
			return err
		}
		trackLog.Name = name
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(csv2gpxCmd)
	csv2gpxCmd.Flags().StringVarP(&csv2gpxTimeFormat, "time-format", "", csv2gpxTimeFormat, "CSV header of time format")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxDate, "date", "", csv2gpxDate, "CSV header of date")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxTime, "time", "", csv2gpxTime, "CSV header of time")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxLatitude, "lat", "", csv2gpxLatitude, "CSV header of latitude")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxLongitude, "lon", "", csv2gpxLongitude, "CSV header of longitude")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxElevation, "ele", "", csv2gpxElevation, "CSV header of elevation")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxName, "name", "", csv2gpxName, "CSV header of waypoint name")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxNamePattern, "regexp", "", csv2gpxNamePattern, "Regexp pattern to filter waypoint by name")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxComment, "cmt", "", csv2gpxComment, "CSV header of waypoint comment")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxDescription, "desc", "", csv2gpxDescription, "CSV header of waypoint description")
	csv2gpxCmd.Flags().BoolVarP(&csv2gpxWaypointsOnly, "wpt", "", csv2gpxWaypointsOnly, "Process waypoints only")
	csv2gpxCmd.Flags().BoolVarP(&csv2gpxTWD97, "twd97", "", csv2gpxTWD97, "Convert coordinates from TWD97 to WGS84")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxEasting, "easting", "", csv2gpxEasting, "CSV header of easting; only used when --twd97 is true")
	csv2gpxCmd.Flags().StringVarP(&csv2gpxNorthing, "northing", "", csv2gpxNorthing, "CSV header of northing; only used when --twd97 is true")
}
