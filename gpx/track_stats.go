package gpx

import (
	"math"
	"time"
)

func NewTrackStats() *TrackStats {
	return &TrackStats{
		Distance:      new(float64),
		NanoDuration:  new(int64),
		ElevationGain: new(float64),
		ElevationLoss: new(float64),
	}
}

func (st *TrackStats) AddTime(duration time.Duration) {
	*st.NanoDuration += duration.Nanoseconds()
}

func (st *TrackStats) StartTime() time.Time {
	return time.Unix(0, st.GetNanoTime())
}

func (st *TrackStats) Duration() time.Duration {
	return time.Nanosecond * time.Duration(*st.NanoDuration)
}

func (s *TrackStats) Merge(o *TrackStats) {
	if s.NanoTime == nil && o.NanoTime != nil {
		s.NanoTime = new(int64)
		*s.NanoTime = *o.NanoTime
	}
	*s.Distance += *o.Distance
	*s.NanoDuration += *o.NanoDuration
	*s.ElevationGain += *o.ElevationGain
	*s.ElevationLoss += *o.ElevationLoss
	if o.ElevationMax != nil {
		if s.ElevationMax == nil {
			s.ElevationMax = new(float64)
			*s.ElevationMax = *o.ElevationMax
		} else {
			*s.ElevationMax = math.Max(*s.ElevationMax, *o.ElevationMax)
		}
	}
	if o.ElevationMin != nil {
		if s.ElevationMin == nil {
			s.ElevationMin = new(float64)
			*s.ElevationMin = *o.ElevationMin
		} else {
			*s.ElevationMin = math.Min(*s.ElevationMin, *o.ElevationMin)
		}
	}
}
