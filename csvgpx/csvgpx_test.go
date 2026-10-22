package csvgpx

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/outdoorsafetylab/gpxtoolkit/elevation"
	"github.com/outdoorsafetylab/gpxtoolkit/gpx"

	"google.golang.org/protobuf/proto"
)

// The .gpx files were written by the csv2gpx command before the conversion
// moved into this package, so these pin that the output did not change.
func TestConvertMatchesCommandOutput(t *testing.T) {
	twd97 := func(o *Options) {
		o.TWD97 = true
		o.Name = "點位名稱"
		o.Easting = "TWD97 東距"
		o.Northing = "TWD97北距"
	}
	cases := []struct {
		csv, gpx string
		opts     func(*Options)
	}{
		{"wgs84.csv", "wgs84.gpx", func(*Options) {}},
		{"wgs84.csv", "wgs84-wpt.gpx", func(o *Options) { o.WaypointsOnly = true }},
		{"twd97.csv", "twd97.gpx", twd97},
		{"twd97.csv", "twd97-regexp.gpx", func(o *Options) { twd97(o); o.NamePattern = "^S-" }},
	}
	for _, c := range cases {
		t.Run(c.gpx, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", c.csv))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			opts := DefaultOptions()
			c.opts(opts)
			tracklog, err := Convert(f, opts)
			if err != nil {
				t.Fatal(err)
			}
			tracklog.Name = proto.String(c.csv)
			var got bytes.Buffer
			err = (&gpx.Writer{Creator: "gpxtoolkit", Writer: &got}).Write(tracklog)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join("testdata", c.gpx))
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != string(want) {
				t.Errorf("got:\n%s\nwant:\n%s", got.String(), want)
			}
		})
	}
}

func TestConvertJoinsDateAndTimeInTimeZone(t *testing.T) {
	opts := DefaultOptions()
	opts.Date = "Date"
	opts.Time = "Time"
	opts.TimeFormat = "2006/01/02 15:04"
	opts.TimeZone = "Asia/Taipei"
	tracklog, err := Convert(strings.NewReader("Date,Time,Latitude,Longitude\n2024/05/01,09:30,24.1,121.2\n"), opts)
	if err != nil {
		t.Fatal(err)
	}
	p := tracklog.Tracks[0].Segments[0].Points[0]
	if got, want := p.GetNanoTime(), int64(1714527000)*1e9; got != want {
		t.Errorf("time = %d, want %d (01:30 UTC)", got, want)
	}
}

type fakeService struct {
	calls [][]*elevation.LatLon
	err   error
}

func (s *fakeService) Lookup(points []*elevation.LatLon) ([]*float64, error) {
	s.calls = append(s.calls, points)
	elevs := make([]*float64, len(points))
	for i, p := range points {
		elevs[i] = proto.Float64(p.Lat * 100)
	}
	return elevs, s.err
}

func TestConvertLooksUpElevationForPointsAndWaypoints(t *testing.T) {
	opts := DefaultOptions()
	service := &fakeService{}
	opts.Service = service
	csv := "Name,Latitude,Longitude,Elevation (m)\n,24,121,1\nPeak,25,121,2\n,26,121,3\n"
	tracklog, err := Convert(strings.NewReader(csv), opts)
	if err != nil {
		t.Fatal(err)
	}
	points := tracklog.Tracks[0].Segments[0].Points
	if len(points) != 2 || points[0].GetElevation() != 2400 || points[1].GetElevation() != 2600 {
		t.Errorf("points not looked up: %v", points)
	}
	if len(tracklog.WayPoints) != 1 || tracklog.WayPoints[0].GetElevation() != 2500 {
		t.Errorf("waypoints not looked up: %v", tracklog.WayPoints)
	}
	if len(service.calls) != 2 {
		t.Errorf("Lookup called %d times, want 2", len(service.calls))
	}
}

func TestConvertReturnsLookupError(t *testing.T) {
	opts := DefaultOptions()
	opts.Service = &fakeService{err: errors.New("down")}
	_, err := Convert(strings.NewReader("Latitude,Longitude\n24,121\n"), opts)
	if err == nil || err.Error() != "down" {
		t.Errorf("err = %v, want down", err)
	}
}

func TestConvertErrors(t *testing.T) {
	cases := []struct {
		name, csv string
		opts      func(*Options)
		want      string
	}{
		{"no latitude", "Lat,Longitude\n24,121\n", func(*Options) {}, "can not find 'Latitude'"},
		{"no longitude", "Latitude,Lon\n24,121\n", func(*Options) {}, "can not find 'Longitude'"},
		{"no easting", "Latitude,Longitude,Northing\n24,121,2650000\n", func(o *Options) { o.TWD97 = true }, "can not find 'Easting'"},
		{"no northing", "Easting,Latitude\n250000,24\n", func(o *Options) { o.TWD97 = true }, "can not find 'Northing'"},
		{"bad pattern", "Latitude,Longitude\n24,121\n", func(o *Options) { o.NamePattern = "(" }, "invalid pattern"},
		{"bad zone", "Latitude,Longitude\n24,121\n", func(o *Options) { o.TimeZone = "Nowhere/City" }, "unknown time zone"},
		{"bad latitude", "Latitude,Longitude\n24,121\nx,121\n", func(*Options) {}, "line 3: "},
		{"bad longitude", "Latitude,Longitude\n24,y\n", func(*Options) {}, "line 2: "},
		{"bad easting", "Easting,Northing\nx,2650000\n", func(o *Options) { o.TWD97 = true }, "line 2: "},
		{"bad northing", "Easting,Northing\n250000,y\n", func(o *Options) { o.TWD97 = true }, "line 2: "},
		{"bad elevation", "Latitude,Longitude,Elevation (m)\n24,121,high\n", func(*Options) {}, "line 2: "},
		{"bad time", "Time (UTC),Latitude,Longitude\nnoon,24,121\n", func(*Options) {}, "line 2: "},
		{"ragged row", "Latitude,Longitude\n24\n", func(*Options) {}, "wrong number of fields"},
		{"empty", "", func(*Options) {}, "EOF"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opts := DefaultOptions()
			c.opts(opts)
			_, err := Convert(strings.NewReader(c.csv), opts)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want it to contain %q", err, c.want)
			}
		})
	}
}
