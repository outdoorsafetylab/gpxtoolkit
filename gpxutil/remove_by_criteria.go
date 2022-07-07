package gpxutil

import (
	"gpxtoolkit/gpx"
	"time"
)

func RemoveDistanceLessThan(distance float64) *RemoveByCriteria {
	return &RemoveByCriteria{
		shouldRemove: func(line *line) bool {
			return line.dist < distance
		},
	}
}

func RemoveDurationLessThan(duration time.Duration) *RemoveByCriteria {
	return &RemoveByCriteria{
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
	shouldRemove func(line *line) bool
}

func (r *RemoveByCriteria) Name() string {
	return "Remove Points by Criteria"
}

func (r *RemoveByCriteria) Run(tracklog *gpx.TrackLog) (int, error) {
	n := 0
	for _, t := range tracklog.Tracks {
		for i, seg := range t.Segments {
			num := len(seg.Points)
			removed, err := r.remove(seg)
			if err != nil {
				return 0, err
			}
			n += (num - len(removed.Points))
			t.Segments[i] = removed
		}
	}
	return n, nil
}

func (r *RemoveByCriteria) remove(seg *gpx.Segment) (*gpx.Segment, error) {
	lines := getLines(seg.Points)
	accepted := make([]*line, 0)
	for _, line := range lines {
		if r.shouldRemove(line) {
			continue
		}
		accepted = append(accepted, line)
	}

	return &gpx.Segment{
		Points: joinLines(accepted),
	}, nil
}
