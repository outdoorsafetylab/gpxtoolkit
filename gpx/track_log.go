package gpx

import (
	"io"
	"time"
)

type TrackLogParser interface {
	Parse(r io.Reader) (*TrackLog, error)
}

type TrackLogCorrector interface {
	Correct(r io.Reader, w io.Writer) error
}

func (log *TrackLog) Time() time.Time {
	return time.Unix(0, log.GetNanoTime())
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

func (log *TrackLog) Stat() *TrackStats {
	st := NewTrackStats()
	for _, t := range log.Tracks {
		st.Merge(t.Stat())
	}
	return st
}

func (log *TrackLog) Filter(vt, ht, slope, alpha float64) int {
	n := 0
	for _, t := range log.Tracks {
		n += t.Filter(vt, ht, slope, alpha)
	}
	return n
}

func (log *TrackLog) OutlierFilter() int {
	n := 0
	for _, t := range log.Tracks {
		n += t.OutlierFilter()
	}
	return n
}

func (log *TrackLog) PointCount() int {
	n := 0
	for _, t := range log.Tracks {
		n += t.PointCount()
	}
	return n
}
