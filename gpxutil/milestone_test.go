package gpxutil

import (
	"bytes"
	"gpxtoolkit/gpx"
	"testing"
)

func TestMilestone(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1" creator="foobar" version="1.1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd">
<trk>
	<trkseg>
	<trkpt lat="25.1707179" lon="121.5534371">
	</trkpt>
	<trkpt lat="25.1706354" lon="121.5532494">
	</trkpt>
	<trkpt lat="25.1705286" lon="121.5530777">
	</trkpt>
	<trkpt lat="25.1704995" lon="121.5528953">
	</trkpt>
	<trkpt lat="25.1699023" lon="121.5522355">
	</trkpt>
	<trkpt lat="25.1710190" lon="121.5491188">
	</trkpt>
	<trkpt lat="25.1768837" lon="121.5474880">
	</trkpt>
	</trkseg>
</trk>
</gpx>`
	tracklog, err := gpx.Parse(bytes.NewBuffer([]byte(xml)))
	if err != nil {
		t.Fatal(err)
	}
	milestone := &Milestone{
		Distance: 100,
		MilestoneName: &MilestoneName{
			Template: `printf("%.1fK", dist/1000)`,
		},
	}
	n, err := milestone.Run(tracklog)
	if err != nil {
		t.Fatal(err)
	}
	if n != 11 {
		t.Fatal(n)
	}
}

func TestMilestoneNameValidate(t *testing.T) {
	vars := &MilestoneNameVariables{
		Number:   1,
		Total:    10,
		Distance: 100.0,
	}
	n := &MilestoneName{
		Template: `printf("%.1fK", dist/1000)`,
	}
	val, err := n.Eval(vars)
	if err != nil {
		t.Fatal(err)
	}
	if val != "0.1K" {
		t.Fatal(val)
	}
	n.Template = `printf("%.0fm", dist)`
	val, err = n.Eval(vars)
	if err != nil {
		t.Fatal(err)
	}
	if val != "100m" {
		t.Fatal(val)
	}
	n.Template = `printf("SM400 %02d/%d", num, total)`
	val, err = n.Eval(vars)
	if err != nil {
		t.Fatal(err)
	}
	if val != "SM400 01/10" {
		t.Fatal(val)
	}
}
