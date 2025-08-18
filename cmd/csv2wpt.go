package cmd

import (
	"encoding/csv"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var (
	csv2wptDate        = ""
	csv2wptTime        = "Time (UTC)"
	csv2wptTimeFormat  = "2006-01-02 15:04:05.999"
	csv2wptTimeZone    = "UTC"
	csv2wptLatitude    = "Latitude"
	csv2wptLongitude   = "Longitude"
	csv2wptElevation   = "Elevation (m)"
	csv2wptName        = "Name"
	csv2wptNamePattern = ""
	csv2wptComment     = "Comment"
	csv2wptDescription = "Description"
)

// csv2wptCmd represents the csv2wpt command
var csv2wptCmd = &cobra.Command{
	Use:   "csv2wpt",
	Short: "Convert CSV to GPX waypoints",
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog := &gpx.TrackLog{}
		var reader *csv.Reader
		switch len(files) {
		case 0:
			reader = csv.NewReader(os.Stdin)
		case 1:
			f, err := os.Open(files[0])
			if err != nil {
				return err
			}
			trackLog.Name = proto.String(files[0])
			reader = csv.NewReader(f)
		default:
			return fmt.Errorf("can only convert one CSV file for each time")
		}
		headers, err := reader.Read()
		if err != nil {
			return err
		}
		var namePattern *regexp.Regexp
		if csv2wptNamePattern != "" {
			namePattern, err = regexp.Compile(csv2wptNamePattern)
			if err != nil {
				return fmt.Errorf("invalid pattern for waypoint name: %s", err.Error())
			}
		}
		dateIndex := -1
		timeIndex := -1
		latitudeIndex := -1
		longitudeIndex := -1
		elevationIndex := -1
		nameIndex := -1
		commentIndex := -1
		descriptionIndex := -1
		for i, header := range headers {
			switch header {
			case csv2wptDate:
				if csv2wptDate != "" {
					dateIndex = i
				}
			case csv2wptTime:
				timeIndex = i
			case csv2wptLatitude:
				latitudeIndex = i
			case csv2wptLongitude:
				longitudeIndex = i
			case csv2wptElevation:
				elevationIndex = i
			case csv2wptName:
				nameIndex = i
			case csv2wptComment:
				commentIndex = i
			case csv2wptDescription:
				descriptionIndex = i
			}
		}
		if latitudeIndex < 0 {
			return fmt.Errorf("can not find '%s' in the CSV headers", csv2wptLatitude)
		}
		if longitudeIndex < 0 {
			return fmt.Errorf("can not find '%s' in the CSV headers", csv2wptLongitude)
		}
		if elevationIndex < 0 {
			fmt.Fprintf(os.Stderr, "Skipping elevation due to '%s' is not found in the CSV headers\n", csv2wptElevation)
		}
		tz, err := time.LoadLocation(csv2wptTimeZone)
		if err != nil {
			return err
		}
		waypointsLatLons := make([]*elevation.LatLon, 0)
		waypoints := make([]*gpx.WayPoint, 0)
		for {
			row, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			waypointName := ""
			if nameIndex >= 0 {
				name := strings.Trim(row[nameIndex], " \t")
				if namePattern == nil || namePattern.MatchString(name) {
					waypointName = name
				}
			}
			var tm time.Time
			if timeIndex >= 0 {
				timeString := strings.Trim(row[timeIndex], " \t")
				if dateIndex >= 0 {
					timeString = fmt.Sprintf("%s %s", strings.Trim(row[dateIndex], " \t"), timeString)
					if err != nil {
						return err
					}
				}
				tm, err = time.ParseInLocation(csv2wptTimeFormat, timeString, tz)
				if err != nil {
					return err
				}
			}
			lat, err := strconv.ParseFloat(strings.Trim(row[latitudeIndex], " \t"), 64)
			if err != nil {
				return err
			}
			lon, err := strconv.ParseFloat(strings.Trim(row[longitudeIndex], " \t"), 64)
			if err != nil {
				return err
			}
			if waypointName != "" {
				wpt := &gpx.WayPoint{
					Name:      proto.String(waypointName),
					Latitude:  proto.Float64(lat),
					Longitude: proto.Float64(lon),
				}
				if !tm.IsZero() {
					wpt.NanoTime = proto.Int64(tm.UnixNano())
				}
			}
			point := &gpx.Point{
				Latitude:  proto.Float64(lat),
				Longitude: proto.Float64(lon),
			}
			if !tm.IsZero() {
				point.NanoTime = proto.Int64(tm.UnixNano())
			}
			if elevationIndex >= 0 {
				elev, err := strconv.ParseFloat(strings.Trim(row[elevationIndex], " \t"), 64)
				if err != nil {
					return err
				}
				point.Elevation = proto.Float64(elev)
			}
			if waypointName != "" {
				wpt := &gpx.WayPoint{
					Name:      proto.String(waypointName),
					NanoTime:  point.NanoTime,
					Latitude:  point.Latitude,
					Longitude: point.Longitude,
					Elevation: point.Elevation,
				}
				if commentIndex >= 0 {
					wpt.Comment = proto.String(strings.Trim(row[commentIndex], " \t"))
				}
				if descriptionIndex >= 0 {
					wpt.Description = proto.String(strings.Trim(row[descriptionIndex], " \t"))
				}
				waypoints = append(waypoints, wpt)
				waypointsLatLons = append(waypointsLatLons, &elevation.LatLon{Lat: lat, Lon: lon})
			}
		}
		service := getElevationService()
		if service != nil {
			elevs, err := service.Lookup(waypointsLatLons)
			if err != nil {
				return err
			}
			for i, elev := range elevs {
				waypoints[i].Elevation = elev
			}
		}
		trackLog.WayPoints = waypoints
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(csv2wptCmd)
	csv2wptCmd.Flags().StringVarP(&csv2wptTimeFormat, "time-format", "", csv2wptTimeFormat, "CSV header of time format")
	csv2wptCmd.Flags().StringVarP(&csv2wptDate, "date", "", csv2wptDate, "CSV header of date")
	csv2wptCmd.Flags().StringVarP(&csv2wptTime, "time", "", csv2wptTime, "CSV header of time")
	csv2wptCmd.Flags().StringVarP(&csv2wptLatitude, "lat", "", csv2wptLatitude, "CSV header of latitude")
	csv2wptCmd.Flags().StringVarP(&csv2wptLongitude, "lon", "", csv2wptLongitude, "CSV header of longitude")
	csv2wptCmd.Flags().StringVarP(&csv2wptElevation, "ele", "", csv2wptElevation, "CSV header of elevation")
	csv2wptCmd.Flags().StringVarP(&csv2wptName, "name", "", csv2wptName, "CSV header of waypoint name")
	csv2wptCmd.Flags().StringVarP(&csv2wptNamePattern, "regexp", "", csv2wptNamePattern, "Regexp pattern to filter waypoint by name")
	csv2wptCmd.Flags().StringVarP(&csv2wptComment, "cmt", "", csv2wptComment, "CSV header of waypoint comment")
	csv2wptCmd.Flags().StringVarP(&csv2wptDescription, "desc", "", csv2wptDescription, "CSV header of waypoint description")
}
