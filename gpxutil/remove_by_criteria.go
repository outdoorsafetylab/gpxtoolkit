package gpxutil

import (
	"gpxtoolkit/gpx"
	"log"
	"time"
)

func RemoveDuplicated() *RemoveByCriteria {
	return &RemoveByCriteria{
		distanceFunc: HaversinDistance,
		shouldRemove: func(line *line) bool {
			return line.a.Equals(line.b)
		},
	}
}

func RemoveDistanceLessThan(distance float64) *RemoveByCriteria {
	return &RemoveByCriteria{
		distanceFunc: HaversinDistance,
		shouldRemove: func(line *line) bool {
			ret := line.dist < distance
			if ret {
				log.Printf("dist=%f, threshold=%f", line.dist, distance)
			}
			return ret
		},
	}
}

func RemoveDurationLessThan(duration time.Duration) *RemoveByCriteria {
	return &RemoveByCriteria{
		distanceFunc: HaversinDistance,
		shouldRemove: func(line *line) bool {
			if line.duration == nil {
				return false
			}
			d := *line.duration
			if d == 0 {
				return false
			}
			if d < 0 {
				d = -d
			}
			return d < duration
		},
	}
}

type RemoveByCriteria struct {
	distanceFunc DistanceFunc
	shouldRemove func(line *line) bool
}

func (r *RemoveByCriteria) Name() string {
	return "Remove Points by Criteria"
}

func (r *RemoveByCriteria) Run(tracklog *gpx.TrackLog) (int, error) {
	n := 0
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			num := len(seg.Points)
			if num <= 1 {
				continue
			}
			removed, err := r.remove(seg.Points)
			if err != nil {
				return 0, err
			}
			n += (num - len(removed))
			seg.Points = removed
		}
	}
	return n, nil
}

func (r *RemoveByCriteria) remove(points []*gpx.Point) ([]*gpx.Point, error) {
	lines := getLines(r.distanceFunc, points)
	accepted := make([]*line, 0)
	for _, line := range lines {
		if r.shouldRemove(line) {
			continue
		}
		accepted = append(accepted, line)
	}
	return joinLines(accepted), nil
}
