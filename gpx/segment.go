package gpx

import (
	"fmt"
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

func (s *Segment) Stat(alpha float64) (*TrackStats, error) {
	st := NewTrackStats()
	*st.NumPoints = int64(len(s.Points))
	var filter *AlphaFilter
	for i, b := range s.Points[1:] {
		a := s.Points[i]
		if a.Elevation == nil {
			return nil, fmt.Errorf("missing elevation in point[%d]", i)
		}
		if b.Elevation == nil {
			return nil, fmt.Errorf("missing elevation in point[%d]", i+1)
		}
		if st.NanoTime == nil && b.NanoTime != nil {
			st.NanoTime = b.NanoTime
		}
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
		if filter == nil {
			filter = &AlphaFilter{Alpha: alpha, Value: elev}
		} else {
			delta := filter.Accumulate(elev - filter.Value)
			if delta > 0 {
				*st.ElevationGain += delta
			} else if delta < 0 {
				*st.ElevationLoss += -delta
			}
		}
		dist := a.distanceTo(b)
		*st.Distance += dist
		*st.ElevationDistance += (a.GetElevation() + b.GetElevation()) / 2 * dist
		st.AddTime(b.Time().Sub(a.Time()))
	}
	return st, nil
}

func (s *Segment) BoundingBox() *BoundingBox {
	bbox := &BoundingBox{}
	for _, p := range s.Points {
		if p.Latitude == nil || p.Longitude == nil {
			continue
		}
		bbox.Add(p.GetLatitude(), p.GetLongitude())
	}
	return bbox
}
