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
	csv2gpxDate          = ""
	csv2gpxTime          = "Time (UTC)"
	csv2gpxTimeFormat    = "2006-01-02 15:04:05.999"
	csv2gpxTimeZone      = "UTC"
	csv2gpxLatitude      = "Latitude"
	csv2gpxLongitude     = "Longitude"
	csv2gpxElevation     = "Elevation (m)"
	csv2gpxName          = "Name"
	csv2gpxNamePattern   = ""
	csv2gpxComment       = "Comment"
	csv2gpxDescription   = "Description"
	csv2gpxWaypointsOnly = false
)

// csv2gpxCmd represents the csv2gpx command
var csv2gpxCmd = &cobra.Command{
	Use:   "csv2gpx",
	Short: "Convert CSV to GPX format",
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
		if csv2gpxNamePattern != "" {
			namePattern, err = regexp.Compile(csv2gpxNamePattern)
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
			case csv2gpxDate:
				if csv2gpxDate != "" {
					dateIndex = i
				}
			case csv2gpxTime:
				timeIndex = i
			case csv2gpxLatitude:
				latitudeIndex = i
			case csv2gpxLongitude:
				longitudeIndex = i
			case csv2gpxElevation:
				elevationIndex = i
			case csv2gpxName:
				nameIndex = i
			case csv2gpxComment:
				commentIndex = i
			case csv2gpxDescription:
				descriptionIndex = i
			}
		}
		if latitudeIndex < 0 {
			return fmt.Errorf("can not find '%s' in the CSV headers", csv2gpxLatitude)
		}
		if longitudeIndex < 0 {
			return fmt.Errorf("can not find '%s' in the CSV headers", csv2gpxLongitude)
		}
		if elevationIndex < 0 {
			fmt.Fprintf(os.Stderr, "Skipping elevation due to '%s' is not found in the CSV headers\n", csv2gpxElevation)
		}
		tz, err := time.LoadLocation(csv2gpxTimeZone)
		if err != nil {
			return err
		}
		pointsLatLons := make([]*elevation.LatLon, 0)
		points := make([]*gpx.Point, 0)
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
				tm, err = time.ParseInLocation(csv2gpxTimeFormat, timeString, tz)
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
			} else if !csv2gpxWaypointsOnly {
				points = append(points, point)
				pointsLatLons = append(pointsLatLons, &elevation.LatLon{Lat: lat, Lon: lon})
			}
		}
		service := getElevationService()
		if service != nil {
			elevs, err := service.Lookup(pointsLatLons)
			if err != nil {
				return err
			}
			for i, elev := range elevs {
				points[i].Elevation = elev
			}
			elevs, err = service.Lookup(waypointsLatLons)
			if err != nil {
				return err
			}
			for i, elev := range elevs {
				waypoints[i].Elevation = elev
			}
		}
		trackLog.WayPoints = waypoints
		if !csv2gpxWaypointsOnly {
			trackLog.Tracks = []*gpx.Track{
				{
					Segments: []*gpx.Segment{
						{
							Points: points,
						},
					},
				},
			}
		}
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
}
