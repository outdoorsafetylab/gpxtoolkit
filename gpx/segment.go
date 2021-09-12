package gpx

import (
	"math"
	"time"

	"github.com/montanaflynn/stats"
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
	var prev *Point
	for _, p := range s.Points {
		if st.NanoTime == nil && p.NanoTime != nil {
			st.NanoTime = new(int64)
			*st.NanoTime = *p.NanoTime
		}
		if prev != nil {
			*st.Distance += p.DistanceTo(prev)
			st.AddTime(p.Time().Sub(prev.Time()))
			if p.Elevation != nil && prev.Elevation != nil {
				delta := *p.Elevation - *prev.Elevation
				if delta > 0 {
					*st.ElevationGain += delta
				} else if delta < 0 {
					*st.ElevationLoss += -delta
				}
			}
		}
		if p.Elevation != nil {
			if st.ElevationMax == nil {
				st.ElevationMax = new(float64)
				*st.ElevationMax = *p.Elevation
			} else {
				*st.ElevationMax = math.Max(*st.ElevationMax, *p.Elevation)
			}
			if st.ElevationMin == nil {
				st.ElevationMin = new(float64)
				*st.ElevationMin = *p.Elevation
			} else {
				*st.ElevationMin = math.Min(*st.ElevationMin, *p.Elevation)
			}
		}
		prev = p
	}
	return st
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

func (s *Segment) ThresholdFilter(horizon, vertical, slope float64) int {
	if math.IsNaN(horizon) && math.IsNaN(vertical) && math.IsNaN(slope) {
		return 0
	}
	elevFilter := !math.IsNaN(vertical) || !math.IsNaN(slope)
	points := make([]*Point, 0)
	n := len(s.Points)
	var prev *Point
	for i, p := range s.Points {
		if prev == nil || i == n-1 {
			points = append(points, p)
			prev = p
		} else {
			dist := p.DistanceTo(prev)
			if dist < horizon {
				continue
			}
			if elevFilter {
				if p.Elevation == nil {
					continue
				}
				elev := *p.Elevation - *prev.Elevation
				if slope > 0 && math.Abs(elev)/dist > slope {
					continue
				}
				if math.Abs(elev) < vertical {
					*p.Elevation = *prev.Elevation
				}
			}
			points = append(points, p)
			prev = p
		}
	}
	s.Points = points
	return n - len(s.Points)
}

func (s *Segment) AlphaFilter(alpha float64) int {
	if alpha == 1.0 {
		return 0
	}
	var highest, lowest *Point
	for _, p := range s.Points {
		if p.Elevation == nil {
			continue
		}
		if highest == nil || *p.Elevation > *highest.Elevation {
			highest = p
		}
		if lowest == nil || *p.Elevation < *lowest.Elevation {
			lowest = p
		}
	}
	n := len(s.Points)
	points := make([]*Point, 0)
	var prev *Point
	var elev float64
	for _, p := range s.Points {
		if p.Elevation == nil {
			continue
		}
		if p != highest && p != lowest && prev != nil {
			*p.Elevation = Round(elev + alpha*(*p.Elevation-elev))
		}
		elev = *p.Elevation
		points = append(points, p)
		prev = p
	}
	s.Points = points
	return n - len(s.Points)
}

type candidatePoint struct {
	point         *Point
	speed         float64
	verticalSpeed float64
}

func (s *Segment) OutlierFilter() int {
	moves := []float64{}
	uphills := []float64{}
	downhills := []float64{}
	rises := []float64{}
	drops := []float64{}
	n := len(s.Points)
	candidates := make([]*candidatePoint, 0)
	var prev *Point
	for _, p := range s.Points {
		if p.Elevation == nil {
			continue
		}
		c := &candidatePoint{point: p}
		if prev != nil {
			if p.GetNanoTime() != prev.GetNanoTime() {
				duration := p.Time().Sub(prev.Time())
				dist := p.DistanceTo(prev)
				elev := 0.0
				if p.Elevation != nil && prev.Elevation != nil {
					elev = (*p.Elevation - *prev.Elevation)
				}
				speed := dist * float64(time.Second) / float64(duration)
				if elev > 0 {
					uphills = append(uphills, speed)
				} else if elev < 0 {
					downhills = append(downhills, speed)
				} else {
					moves = append(moves, speed)
				}
				c.speed = dist * float64(time.Second) / float64(duration)
				c.verticalSpeed = elev * float64(time.Second) / float64(duration)
				if c.verticalSpeed > 0 {
					rises = append(rises, c.verticalSpeed)
				} else if c.verticalSpeed < 0 {
					drops = append(drops, c.verticalSpeed)
				}
			}
		}
		prev = p
		candidates = append(candidates, c)
	}
	outliers := []float64{}
	samples := [][]float64{
		moves,
		uphills,
		downhills,
	}
	for _, sample := range samples {
		o, err := stats.QuartileOutliers(sample)
		if err == nil {
			outliers = append(outliers, o.Mild...)
			outliers = append(outliers, o.Extreme...)
		}
	}
	candidates = removeOutlierPoints(candidates, outliers, false)

	outliers = []float64{}
	samples = [][]float64{
		rises,
		drops,
	}
	for _, sample := range samples {
		o, err := stats.QuartileOutliers(sample)
		if err == nil {
			outliers = append(outliers, o.Mild...)
			outliers = append(outliers, o.Extreme...)
		}
	}
	candidates = removeOutlierPoints(candidates, outliers, true)

	s.Points = make([]*Point, len(candidates))
	for i, c := range candidates {
		s.Points[i] = c.point
	}
	return n - len(s.Points)
}

func removeOutlierPoints(points []*candidatePoint, outliers []float64, vertical bool) []*candidatePoint {
	result := make([]*candidatePoint, 0)
	for _, p := range points {
		outlier := false
		for _, o := range outliers {
			var speed float64
			if vertical {
				speed = p.verticalSpeed
			} else {
				speed = p.speed
			}
			if speed == o {
				outlier = true
				break
			}
		}
		if !outlier {
			result = append(result, p)
		}
	}
	return result
}
