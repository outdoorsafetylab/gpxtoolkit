package gpxutil

import (
	"gpxtoolkit/gpx"
	"testing"
)

func TestMilestone(t *testing.T) {
	log, err := gpx.Open("../gpx/tests/2021-05-01-153620.gpx")
	if err != nil {
		t.Fatal(err)
	}
	milestone := &Milestone{
		Distance: 100,
		MilestoneName: &MilestoneName{
			Template: `printf("%.1fK", dist/1000)`,
		},
	}
	n, err := milestone.Run(log)
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
