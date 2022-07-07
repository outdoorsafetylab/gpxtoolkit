package gpxutil

import (
	"gpxtoolkit/gpx"
)

type Deduplicate struct{}

func (c *Deduplicate) Name() string {
	return "Deduplicate"
}

func (c *Deduplicate) Run(tracklog *gpx.TrackLog) (int, error) {
	n := 0
	for _, t := range tracklog.Tracks {
		for i, seg := range t.Segments {
			num := len(seg.Points)
			seg, err := c.remove(seg)
			if err != nil {
				return 0, err
			}
			n += (num - len(seg.Points))
			t.Segments[i] = seg
		}
	}
	return n, nil
}

func (c *Deduplicate) remove(seg *gpx.Segment) (*gpx.Segment, error) {
	res := &gpx.Segment{
		Points: make([]*gpx.Point, 0),
	}
	for i, b := range seg.Points {
		if i == 0 {
			res.Points = append(res.Points, b)
			continue
		}
		a := res.Points[len(res.Points)-1]
		if a.GetLatitude() == b.GetLatitude() && a.GetLongitude() == b.GetLongitude() && a.Time() == b.Time() {
			continue
		}
		res.Points = append(res.Points, b)
	}
	return res, nil
}
