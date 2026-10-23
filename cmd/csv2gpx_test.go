package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCsv2gpx runs the command on the given arguments and returns its stdout
// and stderr. cobra keeps flag values in package variables, so every one the
// command reads is reset to its default first.
func runCsv2gpx(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	t.Setenv("GOOGLE_ELEVATION_API_KEY", "")
	t.Setenv("ELEVATION_URL", "")
	t.Setenv("ELEVATION_TOKEN", "")
	files = nil
	elevationURL, elevationToken, googleElevationAPIKey = "", "", ""
	csv2gpxDate = ""
	csv2gpxTime = "Time (UTC)"
	csv2gpxTimeFormat = "2006-01-02 15:04:05.999"
	csv2gpxLatitude = "Latitude"
	csv2gpxLongitude = "Longitude"
	csv2gpxElevation = "Elevation (m)"
	csv2gpxName = "Name"
	csv2gpxNamePattern = ""
	csv2gpxComment = "Comment"
	csv2gpxDescription = "Description"
	csv2gpxWaypointsOnly = false
	csv2gpxTWD97 = false
	csv2gpxEasting = "Easting"
	csv2gpxNorthing = "Northing"
	rootCmd.SetArgs(append([]string{"csv2gpx"}, args...))
	capture := func(f **os.File) func() string {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		saved := *f
		*f = w
		done := make(chan string)
		go func() {
			data, _ := io.ReadAll(r)
			done <- string(data)
		}()
		return func() string {
			*f = saved
			w.Close()
			return <-done
		}
	}
	stdout := capture(&os.Stdout)
	stderr := capture(&os.Stderr)
	err := rootCmd.Execute()
	return stdout(), stderr(), err
}

func writeCSV(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "in.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// The .gpx files were written by the command before it was last refactored;
// the output must not change.
func TestCsv2gpxMatchesRecordedOutput(t *testing.T) {
	dir := filepath.Join("test_files", "csv2gpx")
	wgs84 := filepath.Join(dir, "wgs84.csv")
	twd97 := []string{"-f", filepath.Join(dir, "twd97.csv"), "--twd97", "--name", "點位名稱", "--easting", "TWD97 東距", "--northing", "TWD97北距"}
	cases := []struct {
		want string
		args []string
	}{
		{"wgs84.gpx", []string{"-f", wgs84}},
		{"wgs84-wpt.gpx", []string{"-f", wgs84, "--wpt"}},
		{"twd97.gpx", twd97},
		{"twd97-regexp.gpx", append(append([]string{}, twd97...), "--regexp", "^S-")},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			got, _, err := runCsv2gpx(t, c.args...)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join(dir, c.want))
			if err != nil {
				t.Fatal(err)
			}
			// The recorded files name the track after the bare file name.
			got = strings.Replace(got, "<name>"+dir+string(filepath.Separator), "<name>", 1)
			if got != string(want) {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestCsv2gpxJoinsDateAndTime(t *testing.T) {
	in := writeCSV(t, "Date,Time,Latitude,Longitude\n2024/05/01,09:30,24.1,121.2\n")
	out, _, err := runCsv2gpx(t, "-f", in, "--date", "Date", "--time", "Time", "--time-format", "2006/01/02 15:04")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<time>2024-05-01T09:30:00Z</time>") {
		t.Errorf("no joined time in:\n%s", out)
	}
}

func TestCsv2gpxWarnsWhenElevationColumnIsMissing(t *testing.T) {
	_, stderr, err := runCsv2gpx(t, "-f", writeCSV(t, "Latitude,Longitude\n24,121\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr, "Skipping elevation due to 'Elevation (m)' is not found in the CSV headers") {
		t.Errorf("stderr = %q", stderr)
	}
	_, stderr, err = runCsv2gpx(t, "-f", writeCSV(t, "Latitude,Longitude,Elevation (m)\n24,121,5\n"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stderr, "Skipping elevation") {
		t.Errorf("warned with the column present: %q", stderr)
	}
}

// elevationServer answers each [lon, lat] with lat*100, or fails.
func elevationServer(t *testing.T, fail bool) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail {
			http.Error(w, "down", 500)
			return
		}
		var points [][]float64
		if err := json.NewDecoder(r.Body).Decode(&points); err != nil {
			t.Error(err)
		}
		elevs := make([]float64, len(points))
		for i, p := range points {
			elevs[i] = p[1] * 100
		}
		json.NewEncoder(w).Encode(elevs)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestCsv2gpxLooksUpElevationForPointsAndWaypoints(t *testing.T) {
	in := writeCSV(t, "Name,Latitude,Longitude,Elevation (m)\n,24,121,1\nPeak,25,121,2\n")
	out, _, err := runCsv2gpx(t, "-f", in, "--elevation-url", elevationServer(t, false))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<ele>2400.000000</ele>", "<ele>2500.000000</ele>"} {
		if !strings.Contains(out, want) {
			t.Errorf("no %s in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "<ele>1.000000</ele>") || strings.Contains(out, "<ele>2.000000</ele>") {
		t.Errorf("the CSV elevations were kept:\n%s", out)
	}
	_, _, err = runCsv2gpx(t, "-f", in, "--elevation-url", elevationServer(t, true))
	if err == nil {
		t.Error("a failed lookup did not fail the command")
	}
}

func TestCsv2gpxErrors(t *testing.T) {
	dir := filepath.Join("test_files", "csv2gpx")
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"no latitude", []string{"-f", filepath.Join(dir, "twd97.csv")}, "can not find 'Latitude'"},
		{"no longitude", []string{"-f", writeCSV(t, "Latitude,Lon\n24,121\n")}, "can not find 'Longitude'"},
		{"no easting", []string{"-f", filepath.Join(dir, "wgs84.csv"), "--twd97"}, "can not find 'Easting'"},
		{"no northing", []string{"-f", writeCSV(t, "Easting,Latitude\n250000,24\n"), "--twd97"}, "can not find 'Northing'"},
		{"bad pattern", []string{"-f", filepath.Join(dir, "wgs84.csv"), "--regexp", "("}, "invalid pattern"},
		{"bad latitude", []string{"-f", writeCSV(t, "Latitude,Longitude\nx,121\n")}, "invalid syntax"},
		{"bad longitude", []string{"-f", writeCSV(t, "Latitude,Longitude\n24,y\n")}, "invalid syntax"},
		{"bad easting", []string{"-f", writeCSV(t, "Easting,Northing\nx,2650000\n"), "--twd97"}, "invalid syntax"},
		{"bad northing", []string{"-f", writeCSV(t, "Easting,Northing\n250000,y\n"), "--twd97"}, "invalid syntax"},
		{"bad elevation", []string{"-f", writeCSV(t, "Latitude,Longitude,Elevation (m)\n24,121,high\n")}, "invalid syntax"},
		{"bad time", []string{"-f", writeCSV(t, "Time (UTC),Latitude,Longitude\nnoon,24,121\n")}, "cannot parse"},
		{"ragged row", []string{"-f", writeCSV(t, "Latitude,Longitude\n24\n")}, "wrong number of fields"},
		{"empty", []string{"-f", writeCSV(t, "")}, "EOF"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := runCsv2gpx(t, c.args...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want it to contain %q", err, c.want)
			}
		})
	}
}
