package gpx

import (
	"io"
	"slices"
	"time"
)

type TrackLogParser interface {
	Parse(r io.Reader) (*TrackLog, error)
}

type TrackLogCorrector interface {
	Correct(r io.Reader, w io.Writer) error
}

func (log *TrackLog) Time() time.Time {
	return time.Unix(0, log.GetNanoTime()).UTC()
}

func (log *TrackLog) Start() *Point {
	for _, t := range log.Tracks {
		p := t.Start()
		if p != nil {
			return p
		}
	}
	return nil
}

func (log *TrackLog) End() *Point {
	n := len(log.Tracks)
	for i := n - 1; i >= 0; i-- {
		t := log.Tracks[i]
		p := t.End()
		if p != nil {
			return p
		}
	}
	return nil
}

func (log *TrackLog) Stat(alpha float64) (*TrackStats, error) {
	st := NewTrackStats()
	*st.NumTracks = int64(len(log.Tracks))
	for _, t := range log.Tracks {
		_st, err := t.Stat(alpha)
		if err != nil {
			return nil, err
		}
		st.Merge(_st)
	}
	return st, nil
}

func (log *TrackLog) BoundingBox() *BoundingBox {
	bbox := &BoundingBox{}
	for _, t := range log.Tracks {
		bbox.Merge(t.BoundingBox())
	}
	for _, p := range log.WayPoints {
		bbox.Add(p.GetLatitude(), p.GetLongitude())
	}
	return bbox
}

func (log *TrackLog) RemoveWayPoints(wpts ...*WayPoint) {
	for _, wpt := range wpts {
		log.removeWayPoint(wpt)
	}
}

func (log *TrackLog) removeWayPoint(wpt *WayPoint) {
	i := -1
	for j, w := range log.WayPoints {
		if w == wpt {
			i = j
			break
		}
	}
	if i >= 0 {
		log.WayPoints = slices.Delete(log.WayPoints, i, i+1)
	}
}
