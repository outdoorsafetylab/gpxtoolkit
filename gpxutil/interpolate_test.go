package gpxutil

import (
	"bytes"
	"gpxtoolkit/gpx"
	"testing"
)

func TestInterpolate(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1" creator="foobar" version="1.1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd">
<trk>
  <trkseg>
    <trkpt lat="24.0208146" lon="121.2768584">
    </trkpt>
    <trkpt lat="23.944451" lon="121.191298">
    </trkpt>
  </trkseg>
</trk>
</gpx>`
	tracklog, err := gpx.Parse(bytes.NewBuffer([]byte(xml)))
	if err != nil {
		t.Fatal(err)
	}
	c := &Interpolate{
		Distance: 100,
	}
	n, err := c.Run(tracklog)
	if err != nil {
		t.Fatal(err)
	}
	if n != 122 {
		t.Fatalf("Unexpected interpolated number: %d", n)
	}
}
