package milestone

import (
	"gpxtoolkit/gpx"
	"testing"
	"text/template"
)

func TestMarker(t *testing.T) {
	log, err := gpx.Open("../gpx/tests/2021-05-01-153620.gpx")
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.New("").Parse(`{{printf "%.1f" .Kilometer}}K`)
	if err != nil {
		t.Fatal(err)
	}
	marker := &Marker{
		Distance:     100,
		NameTemplate: tmpl,
	}
	err = marker.Mark(log)
	if err != nil {
		t.Fatal(err)
	}
	if len(log.WayPoints) != 11 {
		t.Fatal(len(log.WayPoints))
	}
}
