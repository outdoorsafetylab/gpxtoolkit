package cmd

import (
	"encoding/csv"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"
	"gpxtoolkit/twd97"
	"os"

	"github.com/spf13/cobra"
)

var (
	wpt2csvTWD97 = false
	wpt2csvDist  = false
)

// wpt2csvCmd represents the wpt2csv command
var wpt2csvCmd = &cobra.Command{
	Use:   "wpt2csv",
	Args:  cobra.NoArgs,
	Short: "Convert waypoints in GPX to CSV format",
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		service := getElevationService()
		if len(trackLog.WayPoints) > 0 {
			return convertWaypointsToCSV(trackLog.WayPoints, service, wpt2csvDist)
		} else {
			return fmt.Errorf("no waypoints in the GPX file")
		}
	},
}

func convertWaypointsToCSV(waypoints []*gpx.WayPoint, service elevation.Service, dist bool) error {
	w := csv.NewWriter(os.Stdout)
	headers := []string{
		"Waypoint Name",
		"Waypoint Index",
		"Time (UTC)",
		"Latitude",
		"Longitude",
	}
	if wpt2csvTWD97 {
		headers = append(headers, "TWD97:X", "TWD97:Y")
	}
	headers = append(headers,
		"Elevation (m)",
		"Description",
		"Comment",
		"Symbol")
	if service != nil {
		headers = append(headers, "Elevation (Calibrated) (m)")
	}
	if dist {
		headers = append(headers, "Harversine Distance to Next (m)", "Terrain Distance to Next (m)")
	}
	err := w.Write(headers)
	if err != nil {
		return err
	}
	var elevs []*float64
	if service != nil {
		latLons := make([]*elevation.LatLon, len(waypoints))
		for k, p := range waypoints {
			latLons[k] = &elevation.LatLon{
				Lat: p.GetLatitude(),
				Lon: p.GetLongitude(),
			}
		}
		elevs, err = service.Lookup(latLons)
		if err != nil {
			return err
		}
	}
	for i, p := range waypoints {
		values := []string{
			p.GetName(),
			fmt.Sprintf("%d", i),
			p.Time().Format("2006-01-02 15:04:05.999"),
			fmt.Sprintf("%f", p.GetLatitude()),
			fmt.Sprintf("%f", p.GetLongitude()),
		}
		if wpt2csvTWD97 {
			x, y := twd97.FromWGS84(p.GetLongitude(), p.GetLatitude(), false)
			values = append(values,
				fmt.Sprintf("%.0f", x),
				fmt.Sprintf("%.0f", y),
			)
		}
		values = append(values,
			fmt.Sprintf("%f", p.GetElevation()),
			p.GetDescription(),
			p.GetComment(),
			p.GetSymbol())
		if service != nil {
			values = append(values, fmt.Sprintf("%f", *elevs[i]))
		}
		if dist {
			if i != len(waypoints)-1 {
				values = append(values,
					fmt.Sprintf("%.f", gpxutil.HaversinDistance(p.GetPoint(), waypoints[i+1].GetPoint())),
					fmt.Sprintf("%.f", gpxutil.TerrainDistance(p.GetPoint(), waypoints[i+1].GetPoint())),
				)
			} else {
				values = append(values, "", "")
			}
		}
		err := w.Write(values)
		if err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}

func init() {
	rootCmd.AddCommand(wpt2csvCmd)
	wpt2csvCmd.Flags().BoolVarP(&wpt2csvTWD97, "twd97", "", wpt2csvTWD97, "TWD97")
	wpt2csvCmd.Flags().BoolVarP(&wpt2csvDist, "dist", "", wpt2csvDist, "Calculate distance between waypoints")
}
