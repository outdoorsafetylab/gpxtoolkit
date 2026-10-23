package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCsv2gpx runs the command with every flag the cases use set explicitly,
// since cobra keeps flag values in package variables between runs.
func runCsv2gpx(t *testing.T, args ...string) (string, error) {
	t.Helper()
	t.Setenv("GOOGLE_ELEVATION_API_KEY", "")
	t.Setenv("ELEVATION_URL", "")
	t.Setenv("ELEVATION_TOKEN", "")
	files = nil
	defaults := []string{"csv2gpx", "--wpt=false", "--twd97=false", "--regexp=", "--name=Name", "--easting=Easting", "--northing=Northing"}
	rootCmd.SetArgs(append(defaults, args...))
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	runErr := rootCmd.Execute()
	os.Stdout = stdout
	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out), runErr
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
			got, err := runCsv2gpx(t, c.args...)
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

func TestCsv2gpxErrors(t *testing.T) {
	dir := filepath.Join("test_files", "csv2gpx")
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"no latitude", []string{"-f", filepath.Join(dir, "twd97.csv")}, "can not find 'Latitude'"},
		{"no easting", []string{"-f", filepath.Join(dir, "wgs84.csv"), "--twd97"}, "can not find 'Easting'"},
		{"bad pattern", []string{"-f", filepath.Join(dir, "wgs84.csv"), "--regexp", "("}, "invalid pattern"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := runCsv2gpx(t, c.args...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want it to contain %q", err, c.want)
			}
		})
	}
}
