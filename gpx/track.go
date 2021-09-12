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

func (t *Track) Stat() *TrackStats {
	st := NewTrackStats()
	for _, s := range t.Segments {
		st.Merge(s.Stat())
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

func (t *Track) Filter(vt, ht, slope, alpha float64) int {
	n := 0
	for _, s := range t.Segments {
		n += s.ThresholdFilter(ht, vt, slope)
		n += s.AlphaFilter(alpha)
	}
	return n
}

func (t *Track) OutlierFilter() int {
	n := 0
	for _, s := range t.Segments {
		n += s.OutlierFilter()
	}
	return n
}

func (t *Track) PointCount() int {
	n := 0
	for _, s := range t.Segments {
		n += len(s.Points)
	}
	return n
}

func (t *Track) BoundingBox() *BoundingBox {
	bbox := &BoundingBox{}
	for _, s := range t.Segments {
		bbox.Merge(s.BoundingBox())
	}
	return bbox
}
