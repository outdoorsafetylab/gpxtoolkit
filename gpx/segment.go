package gpx

import (
	"math"

	"google.golang.org/protobuf/proto"
)

func (s *Segment) Start() *Point {
	if len(s.Points) > 0 {
		return s.Points[0]
	} else {
		return nil
	}
}

func (s *Segment) End() *Point {
	n := len(s.Points)
	if n > 0 {
		return s.Points[n-1]
	} else {
		return nil
	}
}

func (s *Segment) Stat() *TrackStats {
	st := NewTrackStats()
	*st.NumPoints = int64(len(s.Points))
	for i, b := range s.Points {
		if st.NanoTime == nil && b.NanoTime != nil {
			st.NanoTime = b.NanoTime
		}
		if b.Elevation != nil {
			elev := b.GetElevation()
			if st.ElevationMax == nil {
				st.ElevationMax = proto.Float64(elev)
			} else {
				st.ElevationMax = proto.Float64(math.Max(st.GetElevationMax(), elev))
			}
			if st.ElevationMin == nil {
				st.ElevationMin = proto.Float64(elev)
			} else {
				st.ElevationMin = proto.Float64(math.Min(st.GetElevationMin(), elev))
			}
		}
		if i == 0 {
			continue
		}
		a := s.Points[i-1]
		*st.Distance += a.DistanceTo(b)
		st.AddTime(b.Time().Sub(a.Time()))
		if a.Elevation != nil && b.Elevation != nil {
			delta := b.GetElevation() - a.GetElevation()
			if delta > 0 {
				*st.ElevationGain += delta
			} else if delta < 0 {
				*st.ElevationLoss += -delta
			}
		}
	}
	return st
}
