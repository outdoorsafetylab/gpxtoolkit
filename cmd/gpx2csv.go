package cmd

import (
	"encoding/csv"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"
	"gpxtoolkit/twd97"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	gpx2csvTWD97 = false
)

// gpx2csvCmd represents the gpx2csv command
var gpx2csvCmd = &cobra.Command{
	Use:   "gpx2csv",
	Args:  cobra.NoArgs,
	Short: "Convert GPX to CSV format",
	RunE: func(cmd *cobra.Command, args []string) error {
		trackLog, err := loadGpx()
		if err != nil {
			return err
		}
		service := getElevationService()
		if len(trackLog.Tracks) > 0 {
			return convertTracksToCSV(trackLog.Tracks, service)
		} else {
			return fmt.Errorf("no tracks in the GPX file")
		}
	},
}

func convertTracksToCSV(tracks []*gpx.Track, service elevation.Service) error {
	w := csv.NewWriter(os.Stdout)
	headers := []string{
		"Track Name",
		"Track Index",
		"Segment Index",
		"Point Index",
		"Time (UTC)",
		"Latitude",
		"Longitude",
		"Elevation (m)",
		"Elapsed Time (H:M:S)",
		"Elapsed Time (sec)",
		"Mileage (m)",
		"Leg Time (H:M:S)",
		"Leg Time (sec)",
		"Leg Distance (m)",
		"Leg Speed (Km/H)",
		"Gain/Loss (m)",
		"Slope",
		"Vertical Speed (m/H)",
		"KmE/H",
		"EpH",
	}
	if service != nil {
		headers = append(headers,
			"Elevation (Calibrated) (m)",
			"Gain/Loss (Calibrated) (m)",
			"Slope (Calibrated)",
			"Vertical Speed (Calibrated) (m/H)",
			"KmE/H (Calibrated)",
			"EpH (Calibrated)")
	}
	if gpx2csvTWD97 {
		headers = append(headers,
			"TWD97 TM2 X (m)",
			"TWD97 TM2 Y (m)")
	}
	err := w.Write(headers)
	if err != nil {
		return err
	}
	for i, t := range tracks {
		for j, s := range t.Segments {
			var elevs []*float64
			if service != nil {
				latLons := make([]*elevation.LatLon, len(s.Points))
				for k, p := range s.Points {
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
			mileage := 0.0
			duration := time.Duration(0)
			for k, p := range s.Points {
				dist := 0.0
				var dt time.Duration
				hspeed := 0.0
				elevGainLoss := 0.0
				elevGainLoss1 := 0.0
				slope := 0.0
				slope1 := 0.0
				vspeed := 0.0
				vspeed1 := 0.0
				KmEpH := 0.0
				KmEpH1 := 0.0
				EpH := 0.0
				EpH1 := 0.0
				if k > 0 {
					p0 := s.Points[k-1]
					dist = gpxutil.HaversinDistance(p0, p)
					dt = p.Time().Sub(p0.Time())
					hspeed = (dist / 1000.0) / float64(dt.Hours())
					elevGainLoss = p.GetElevation() - p0.GetElevation()
					slope = elevGainLoss / dist
					vspeed = elevGainLoss / float64(dt.Hours())
					KmE := dist / 1000.0
					KmE1 := KmE
					Ep := KmE
					Ep1 := KmE
					if elevGainLoss >= 0 {
						KmE += elevGainLoss / 100.0
						Ep += elevGainLoss / 100.0
					} else {
						KmE -= elevGainLoss / 333.3
					}
					if service != nil {
						elevGainLoss1 = *elevs[k] - *elevs[k-1]
						slope1 = elevGainLoss1 / dist
						vspeed1 = elevGainLoss1 / float64(dt.Hours())
						if elevGainLoss1 >= 0 {
							KmE1 += elevGainLoss1 / 100.0
							Ep1 += elevGainLoss1 / 100.0
						} else {
							KmE1 -= elevGainLoss1 / 333.3
						}
					}
					KmEpH = KmE / float64(dt.Hours())
					KmEpH1 = KmE1 / float64(dt.Hours())
					EpH = Ep / float64(dt.Hours())
					EpH1 = Ep1 / float64(dt.Hours())
					duration += dt
				}
				mileage += dist
				values := []string{
					t.GetName(),
					fmt.Sprintf("%d", i),
					fmt.Sprintf("%d", j),
					fmt.Sprintf("%d", k),
					p.Time().Format("2006-01-02 15:04:05.999"),
					fmt.Sprintf("%f", p.GetLatitude()),
					fmt.Sprintf("%f", p.GetLongitude()),
					fmt.Sprintf("%f", p.GetElevation()),
					fmt.Sprintf("%v", durationToHMSs(duration)),
					fmt.Sprintf("%f", duration.Seconds()),
					fmt.Sprintf("%f", mileage),
					fmt.Sprintf("%v", durationToHMSs(dt)),
					fmt.Sprintf("%f", dt.Seconds()),
					fmt.Sprintf("%f", dist),
					fmt.Sprintf("%f", hspeed),
					fmt.Sprintf("%f", elevGainLoss),
					fmt.Sprintf("%f", slope),
					fmt.Sprintf("%f", vspeed),
					fmt.Sprintf("%f", KmEpH),
					fmt.Sprintf("%f", EpH),
				}
				if service != nil {
					values = append(values,
						fmt.Sprintf("%f", *elevs[k]),
						fmt.Sprintf("%f", elevGainLoss1),
						fmt.Sprintf("%f", slope1),
						fmt.Sprintf("%f", vspeed1),
						fmt.Sprintf("%f", KmEpH1),
						fmt.Sprintf("%f", EpH1))
				}
				if gpx2csvTWD97 {
					// Convert WGS84 coordinates to TWD97 TM2
					x, y := twd97.FromWGS84(p.GetLongitude(), p.GetLatitude(), false)
					values = append(values,
						fmt.Sprintf("%.2f", x),
						fmt.Sprintf("%.2f", y))
				}
				err := w.Write(values)
				if err != nil {
					return err
				}
			}
		}
	}
	w.Flush()
	return nil
}

func init() {
	rootCmd.AddCommand(gpx2csvCmd)
	gpx2csvCmd.Flags().BoolVarP(&gpx2csvTWD97, "twd97", "", gpx2csvTWD97, "Add TWD97 coordinates")
}

func durationToHMSs(d time.Duration) string {
	millis := d.Milliseconds()
	seconds := millis / 1000
	millis %= 1000
	minutes := seconds / 60
	seconds %= 60
	hours := minutes / 60
	minutes %= 60
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)
}
