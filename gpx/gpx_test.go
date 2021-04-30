package gpx

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
)

func TestInvalidGPX(t *testing.T) {
	p := &Parser{}
	_, err := p.Parse(bytes.NewBuffer([]byte("abc")))
	if err == nil {
		t.Fatalf("Should return err")
	}
	_, err = p.Parse(bytes.NewBuffer([]byte(``)))
	if err == nil {
		t.Fatalf("Should return err")
	}
}

func TestEmptyGPX(t *testing.T) {
	p := &Parser{}
	log, err := p.Parse(bytes.NewBuffer([]byte(`<gpx></gpx>`)))
	if err != nil {
		t.Fatal(err)
	}
	if log == nil {
		t.Fatalf("Should not be nil")
	}
	if log.Creator != nil {
		t.Fatalf("Should not has creator")
	}
	if log.Name != nil {
		t.Fatalf("Should not has name")
	}
	if len(log.Tracks) != 0 {
		t.Fatalf("Should not has track")
	}
}

func TestSimpleGPX(t *testing.T) {
	p := &Parser{}
	creator := "test_creator"
	name := "test_name"
	linkUrl := "http://foobar"
	linkText := "foobar"
	waypointName := "test_waypoint"
	trackName := "test_track"
	trackComment := "test_track_cmt"
	start := time.Now()
	gpx := fmt.Sprintf(`<gpx creator="%s">
		<metadata>
		<name>%s</name>
		<link href="%s"><text>%s</text></link>
		</metadata>
		<wpt lat="1" lon="1"><name>%s1</name><time>%s</time><ele>1001</ele></wpt>
		<wpt lat="2" lon="2"><name>%s2</name><time>%s</time><ele>2002</ele></wpt>
		<trk>
		<name>%s1</name>
		<cmt>%s1</cmt>
		<trkseg>
		<trkpt lat="1" lon="1"><time>%s</time><ele>1001</ele></trkpt>
		<trkpt lat="1" lon="2"><time>%s</time><ele>1002</ele></trkpt>
		</trkseg>
		<trkseg>
		<trkpt lat="2" lon="1"><time>%s</time><ele>2001</ele></trkpt>
		<trkpt lat="2" lon="2"><time>%s</time><ele>2002</ele></trkpt>
		</trkseg>
		</trk>
		<trk>
		<name>%s2</name>
		<cmt>%s2</cmt>
		<trkseg>
		<trkpt lat="3" lon="1"><time>%s</time><ele>3001</ele></trkpt>
		<trkpt lat="3" lon="2"><time>%s</time><ele>3002</ele></trkpt>
		</trkseg>
		<trkseg>
		<trkpt lat="4" lon="1"><time>%s</time><ele>4001</ele></trkpt>
		<trkpt lat="4" lon="2"><time>%s</time><ele>4002</ele></trkpt>
		</trkseg>
		</trk>
		</gpx>`,
		creator, name, linkUrl, linkText,
		waypointName, start.Add(1*time.Second).Format(time.RFC3339),
		waypointName, start.Add(2*time.Second).Format(time.RFC3339),
		trackName, trackComment,
		start.Add(1*time.Second).Format(time.RFC3339),
		start.Add(2*time.Second).Format(time.RFC3339),
		start.Add(3*time.Second).Format(time.RFC3339),
		start.Add(4*time.Second).Format(time.RFC3339),
		trackName, trackComment,
		start.Add(5*time.Second).Format(time.RFC3339),
		start.Add(6*time.Second).Format(time.RFC3339),
		start.Add(7*time.Second).Format(time.RFC3339),
		start.Add(8*time.Second).Format(time.RFC3339),
	)
	log, err := p.Parse(bytes.NewBuffer([]byte(gpx)))
	if err != nil {
		t.Fatal(err)
	}
	if log == nil {
		t.Fatalf("Should not be nil")
	}
	if log.Creator == nil || *log.Creator != creator {
		t.Fatalf("Mismatched creator")
	}
	if log.Name == nil || *log.Name != name {
		t.Fatalf("Mismatched name")
	}
	if log.Link == nil || log.Link.Url == nil || log.Link.Text == nil || *log.Link.Url != linkUrl || *log.Link.Text != linkText {
		t.Fatalf("Mismatched link")
	}
	if len(log.WayPoints) != 2 {
		t.Fatalf("Mismatched num waypoints")
		for i, wpt := range log.WayPoints {
			num := i + 1
			if wpt.GetName() != fmt.Sprintf("%s%d", waypointName, num) {
				t.Fatalf("Mismatched track name")
			}
			if wpt.GetLatitude() != float64(num) {
				t.Fatalf("Mismatched wpt lat")
			}
			if wpt.GetLongitude() != float64(num) {
				t.Fatalf("Mismatched wpt lon")
			}
			if wpt.GetElevation() != wpt.GetLatitude()*1000+wpt.GetLongitude() {
				t.Fatalf("Mismatched wpt ele")
			}
		}
	}
	if len(log.Tracks) != 2 {
		t.Fatalf("Mismatched num tracks")
		num := 1
		numPt := 1
		for _, tr := range log.Tracks {
			if tr.GetName() != fmt.Sprintf("%s%d", trackName, num) {
				t.Fatalf("Mismatched track name")
			}
			if tr.GetComment() != fmt.Sprintf("%s%d", trackComment, num) {
				t.Fatalf("Mismatched track cmt")
			}
			if len(tr.Segments) != 1 {
				t.Fatalf("Mismatched num segments")
				for i, pt := range tr.Segments[0].Points {
					if pt.GetLatitude() != float64(numPt) {
						t.Fatalf("Mismatched pt lat")
					}
					if pt.GetLongitude() != float64(i+1) {
						t.Fatalf("Mismatched pt lon")
					}
					if pt.GetElevation() != pt.GetLatitude()*1000+pt.GetLongitude() {
						t.Fatalf("Mismatched pt ele")
					}
					numPt++
				}
			}
			num++
		}
	}
	var buf bytes.Buffer
	w := &Writer{
		Creator: creator,
		Writer:  &buf,
	}
	err = w.Write(log)
	if err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	log2, err := p.Parse(&buf)
	if err != nil {
		t.Fatal(err)
	}
	data1, err := proto.Marshal(log)
	if err != nil {
		t.Fatal(err)
	}
	data2, err := proto.Marshal(log2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data1, data2) {
		t.Fatalf("Mismatched data:\n%s\n%s", gpx, data)
	}
}

func TestGPXFile(t *testing.T) {
	path := "tests/hiking_1f18be7c8a5c5f62fac3cd5c0d46b648.gpx"
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	p := &Parser{}
	log, err := p.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(log.WayPoints) != 4 {
		t.Fatalf("Unexpected number of way points in %s: %d", path, len(log.WayPoints))
	}
	names := []string{
		"夢幻湖國家級濕地",
		"七星山東峰三角點",
		"七星山主峰",
		"冷水坑登山口",
	}
	for i, name := range names {
		if log.WayPoints[i].GetName() != name {
			t.Fatalf("Unexpected name of way point %d: %s", i, log.WayPoints[i].GetName())
		}
	}
	tracks := log.Tracks
	if len(tracks) != 4 {
		t.Fatalf("Unexpected number of tracks in %s: %d", path, len(tracks))
	}
	names = []string{
		"Track 0 (12:39:57)",
		"Track 1 (13:08:29)",
		"Track 2 (14:06:47)",
		"Track 3 (14:39:37)",
	}
	for i, name := range names {
		if tracks[i].GetName() != name {
			t.Fatalf("Unexpected name of track %d: %s", i, tracks[i].GetName())
		}
	}
	for i, track := range tracks {
		if len(track.Segments) != 1 {
			t.Fatalf("Unexpected number of segments in track %d: %d", i, len(track.Segments))
		}
		seg := track.Segments[0]
		switch i {
		case 0:
			for j, pt := range seg.Points {
				switch j {
				case 0:
					if pt.GetLatitude() != 25.167297 {
						t.Fatalf("Unexpected latitude of track %d's point %d: %f", i, j, pt.GetLatitude())
					}
					if pt.GetLongitude() != 121.562974 {
						t.Fatalf("Unexpected longitude of track %d's point %d: %f", i, j, pt.GetLongitude())
					}
					if pt.Elevation == nil {
						t.Fatalf("Missing elevation of track %d's point %d", i, j)
					}
					if *pt.Elevation != 741.136687 {
						t.Fatalf("Unexpected elevation of track %d's point %d: %f", i, j, *pt.Elevation)
					}
					tm, _ := time.Parse(time.RFC3339, "2018-11-15T04:39:57.999Z")
					if !pt.Time().Equal(tm) {
						t.Fatalf("Unexpected time of track %d's point %d: %v vs %v", i, j, pt.Time(), tm)
					}
				}
			}
		}
	}
}

/*
func TestGPXCorrect(t *testing.T) {
	path := "tests/hiking_1f18be7c8a5c5f62fac3cd5c0d46b648.gpx"
	in, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out := &bytes.Buffer{}
	c := &GpxCorrector{
		Service: &MOIDEMService{
			Client: http.DefaultClient,
			URL:    "http://127.0.0.1:8081",
		},
	}
	err = c.Correct(in, out)
	if err != nil {
		t.Fatal(err)
	}
	path = fmt.Sprintf("%s.out", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = ioutil.WriteFile(path, out.Bytes(), 0644)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(data, out.Bytes()) != 0 {
			t.Fatalf("Unexpected result of corrected GPX: %s", path)
		}
	}
}
*/
