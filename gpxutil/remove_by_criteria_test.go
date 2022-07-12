package gpxutil

import (
	"bytes"
	"gpxtoolkit/gpx"
	"testing"
)

func TestRemoveDuplicatedLatLon(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1" creator="foobar" version="1.1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd">
	<trk>
	<trkseg>
		<trkpt lat="25.1707179" lon="121.5534371">
		</trkpt>
		<trkpt lat="25.1707179" lon="121.5534371">
		</trkpt>
		<trkpt lat="25.1706354" lon="121.5532494">
		</trkpt>
	</trkseg>
	</trk>
</gpx>`
	tracklog, err := gpx.Parse(bytes.NewBuffer([]byte(xml)))
	if err != nil {
		t.Fatal(err)
	}
	n, err := RemoveDuplicated().Run(tracklog)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("Unexpected removed number: %d", n)
	}
}

func TestRemoveDuplicatedLatLonAndTime(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1" creator="foobar" version="1.1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd">
	<trk>
	<trkseg>
		<trkpt lat="25.1707179" lon="121.5534371">
			<time>2018-10-28T10:19:57Z</time>
		</trkpt>
		<trkpt lat="25.1707179" lon="121.5534371">
			<time>2018-10-28T10:19:57Z</time>
		</trkpt>
		<trkpt lat="25.1707179" lon="121.5534371">
			<time>2018-10-28T10:19:58Z</time>
		</trkpt>
	</trkseg>
	</trk>
</gpx>`
	tracklog, err := gpx.Parse(bytes.NewBuffer([]byte(xml)))
	if err != nil {
		t.Fatal(err)
	}
	n, err := RemoveDuplicated().Run(tracklog)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("Unexpected removed number: %d", n)
	}
}

func TestRemoveDuplicatedLatLonEleAndTime(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1" creator="foobar" version="1.1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd">
	<trk>
	<trkseg>
		<trkpt lat="25.1707179" lon="121.5534371">
			<time>2018-10-28T10:19:57Z</time>
			<ele>1100</ele>
		</trkpt>
		<trkpt lat="25.1707179" lon="121.5534371">
			<time>2018-10-28T10:19:57Z</time>
			<ele>1100</ele>
		</trkpt>
		<trkpt lat="25.1707179" lon="121.5534371">
			<time>2018-10-28T10:19:57Z</time>
			<ele>1101</ele>
		</trkpt>
	</trkseg>
	</trk>
</gpx>`
	tracklog, err := gpx.Parse(bytes.NewBuffer([]byte(xml)))
	if err != nil {
		t.Fatal(err)
	}
	n, err := RemoveDuplicated().Run(tracklog)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("Unexpected removed number: %d", n)
	}
}

func TestRemoveDuplicatedRealData(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1" creator="foobar" version="1.1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd">
	<trk>
	<trkseg>
		<trkpt lat="23.481811" lon="120.886987">
			<ele>2611.449951</ele>
			<time>2020-10-26T06:57:17Z</time>
		</trkpt>
		<trkpt lat="23.478817" lon="120.904876">
			<ele>2611.449951</ele>
			<time>2020-10-26T06:59:03Z</time>
		</trkpt>
		<trkpt lat="23.478817" lon="120.904876">
			<ele>2611.449951</ele>
			<time>2020-10-26T06:59:03Z</time>
		</trkpt>
		<trkpt lat="23.481703" lon="120.886909">
			<ele>2622.003430</ele>
			<time>2020-10-26T06:59:09Z</time>
		</trkpt>
	</trkseg>
	</trk>
</gpx>`
	tracklog, err := gpx.Parse(bytes.NewBuffer([]byte(xml)))
	if err != nil {
		t.Fatal(err)
	}
	n, err := RemoveDuplicated().Run(tracklog)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("Unexpected removed number: %d", n)
	}
}
