package gpx

func (t *Track) Start() *Point {
	for _, s := range t.Segments {
		p := s.Start()
		if p != nil {
			return p
		}
	}
	return nil
}

func (t *Track) End() *Point {
	n := len(t.Segments)
	for i := n - 1; i >= 0; i-- {
		s := t.Segments[i]
		p := s.End()
		if p != nil {
			return p
		}
	}
	return nil
}

func (t *Track) Stat(alpha float64) *TrackStats {
	st := NewTrackStats()
	*st.NumSegments = int64(len(t.Segments))
	for _, s := range t.Segments {
		st.Merge(s.Stat(alpha))
	}
	return st
}

func (t *Track) Points() []*Point {
	points := make([]*Point, 0)
	for _, s := range t.Segments {
		points = append(points, s.Points...)
	}
	return points
}
