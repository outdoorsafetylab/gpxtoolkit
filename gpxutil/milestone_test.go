package gpxutil

import (
	"bytes"
	"github.com/outdoorsafetylab/gpxtoolkit/gpx"
	"math"
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

func TestMilestoneNameRejectsWideFormats(t *testing.T) {
	ok := []string{
		`printf("%.1fK", dist/1000)`,
		`printf("%03d/%d", num, total)`,
		`printf("%99s|%-5.2f%%", "a", elev)`,
		`printf("%10.10f", dist)`,
		`printf("km 100: %d", num)`,
		`printf("%+-# 099.99e|%c|%q|%t|%x|%%", dist, 65, "a", true, 255)`,
	}
	for _, template := range ok {
		name := &MilestoneName{Template: template}
		if _, err := name.Eval(&MilestoneNameVariables{Number: 1, Total: 2}); err != nil {
			t.Errorf("%s: %v", template, err)
		}
	}
	bad := []string{
		`printf("%100s", "")`,
		`printf("%.1000000f", dist)`,
		`printf("%" + "1000000" + "s", "")`,
		`printf("%*s", num, "")`,
		`printf("%1000000d", 7)`,
		`printf("%.*f", num, dist)`,
		`printf("%_%1000000s", "", "")`,
		`printf("%_%*s", "", num, "")`,
		`printf("%é%1000s", "")`,
		`printf("%[1]1000000s", "")`,
		`printf("%1-%1000000s", "", "")`,
		`printf("%.0.%1000000s", "", "")`,
		`printf("%1-%*s", "", num, "")`,
		`printf("%1000000T", num)`,
		`printf("%1000000p", num)`,
		`printf("%T", num)`,
		`printf("%[2]d-%[1]d", num, total)`,
		`printf("%1000000%")`,
		`printf()`,
	}
	for _, template := range bad {
		name := &MilestoneName{Template: template}
		if _, err := name.Eval(&MilestoneNameVariables{Number: 1, Total: 2}); err == nil {
			t.Errorf("%s: no error", template)
		}
	}
}

func TestMilestoneRejectsInvalidDistance(t *testing.T) {
	for _, d := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		tracklog := &gpx.TrackLog{}
		m := &Milestone{Distance: d, MilestoneName: &MilestoneName{Template: `"x"`}}
		if _, err := m.Run(tracklog); err == nil {
			t.Errorf("distance %v: no error", d)
		}
	}
}
