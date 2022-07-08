package gpxutil

import (
	"gpxtoolkit/gpx"
	"testing"
	"text/template"
)

func TestMilestone(t *testing.T) {
	log, err := gpx.Open("../gpx/tests/2021-05-01-153620.gpx")
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.New("").Parse(`{{printf "%.1f" .Kilometer}}K`)
	if err != nil {
		t.Fatal(err)
	}
	milestone := &Milestone{
		Distance: 100,
		Template: tmpl,
	}
	n, err := milestone.Run(log)
	if err != nil {
		t.Fatal(err)
	}
	if n != 11 {
		t.Fatal(n)
	}
}
