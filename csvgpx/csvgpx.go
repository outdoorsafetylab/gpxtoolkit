// Package csvgpx converts a CSV table of coordinates into a GPX track log.
package csvgpx

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/outdoorsafetylab/gpxtoolkit/elevation"
	"github.com/outdoorsafetylab/gpxtoolkit/gpx"
	"github.com/outdoorsafetylab/gpxtoolkit/log"
	"github.com/outdoorsafetylab/gpxtoolkit/twd97"

	"google.golang.org/protobuf/proto"
)

// Options names the CSV columns to read and how to interpret them.
type Options struct {
	Date          string // header of the date column; empty when the time column carries the date
	Time          string
	TimeFormat    string
	TimeZone      string
	Latitude      string
	Longitude     string
	Elevation     string
	Name          string
	NamePattern   string // regexp; rows whose name does not match become track points
	Comment       string
	Description   string
	WaypointsOnly bool
	TWD97         bool // read Easting/Northing in TWD97 instead of Latitude/Longitude
	Easting       string
	Northing      string
	// Service looks up elevations for every point when set, replacing the
	// elevation column.
	Service elevation.Service
}

// DefaultOptions returns the column names the csv2gpx command uses by default.
func DefaultOptions() *Options {
	return &Options{
		Time:        "Time (UTC)",
		TimeFormat:  "2006-01-02 15:04:05.999",
		TimeZone:    "UTC",
		Latitude:    "Latitude",
		Longitude:   "Longitude",
		Elevation:   "Elevation (m)",
		Name:        "Name",
		Comment:     "Comment",
		Description: "Description",
		Easting:     "Easting",
		Northing:    "Northing",
	}
}

// Convert reads the CSV from r. A row with a name becomes a waypoint, any
// other row a point of a single track segment.
func Convert(r io.Reader, opts *Options) (*gpx.TrackLog, error) {
	reader := csv.NewReader(r)
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	var namePattern *regexp.Regexp
	if opts.NamePattern != "" {
		namePattern, err = regexp.Compile(opts.NamePattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern for waypoint name: %s", err.Error())
		}
	}
	dateIndex := -1
	timeIndex := -1
	latitudeIndex := -1
	longitudeIndex := -1
	eastingIndex := -1
	northingIndex := -1
	elevationIndex := -1
	nameIndex := -1
	commentIndex := -1
	descriptionIndex := -1
	for i, header := range headers {
		switch header {
		case opts.Date:
			if opts.Date != "" {
				dateIndex = i
			}
		case opts.Time:
			timeIndex = i
		case opts.Latitude:
			latitudeIndex = i
		case opts.Longitude:
			longitudeIndex = i
		case opts.Easting:
			eastingIndex = i
		case opts.Northing:
			northingIndex = i
		case opts.Elevation:
			elevationIndex = i
		case opts.Name:
			nameIndex = i
		case opts.Comment:
			commentIndex = i
		case opts.Description:
			descriptionIndex = i
		}
	}
	if opts.TWD97 {
		if eastingIndex < 0 {
			return nil, fmt.Errorf("can not find '%s' in the CSV headers", opts.Easting)
		}
		if northingIndex < 0 {
			return nil, fmt.Errorf("can not find '%s' in the CSV headers", opts.Northing)
		}
	} else {
		if latitudeIndex < 0 {
			return nil, fmt.Errorf("can not find '%s' in the CSV headers", opts.Latitude)
		}
		if longitudeIndex < 0 {
			return nil, fmt.Errorf("can not find '%s' in the CSV headers", opts.Longitude)
		}
	}
	if elevationIndex < 0 {
		log.Infof("Skipping elevation due to '%s' is not found in the CSV headers", opts.Elevation)
	}
	tz, err := time.LoadLocation(opts.TimeZone)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		line, _ := reader.FieldPos(0)
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
			}
			tm, err = time.ParseInLocation(opts.TimeFormat, timeString, tz)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}
		}
		var lat, lon float64
		if opts.TWD97 {
			x, err := parseFloat(row[eastingIndex])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}
			y, err := parseFloat(row[northingIndex])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}
			lat, lon = twd97.ToWGS84(x, y, false)
		} else {
			lat, err = parseFloat(row[latitudeIndex])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}
			lon, err = parseFloat(row[longitudeIndex])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
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
			elev, err := parseFloat(row[elevationIndex])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
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
		} else if !opts.WaypointsOnly {
			points = append(points, point)
			pointsLatLons = append(pointsLatLons, &elevation.LatLon{Lat: lat, Lon: lon})
		}
	}
	if opts.Service != nil {
		elevs, err := opts.Service.Lookup(pointsLatLons)
		if err != nil {
			return nil, err
		}
		for i, elev := range elevs {
			points[i].Elevation = elev
		}
		elevs, err = opts.Service.Lookup(waypointsLatLons)
		if err != nil {
			return nil, err
		}
		for i, elev := range elevs {
			waypoints[i].Elevation = elev
		}
	}
	trackLog := &gpx.TrackLog{
		WayPoints: waypoints,
	}
	if len(points) > 0 {
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
	return trackLog, nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Trim(s, " \t"), 64)
}
