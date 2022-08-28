package gpxutil

import (
	"fmt"
	"math"

	"gpxtoolkit/gpx"
	"gpxtoolkit/log"
)

func (c *RemoveOutlier) Name() string {
	return fmt.Sprintf("Remove Outliers by %s", c.metric)
}

func RemoveOutlierBySpeed(sigma int) *RemoveOutlier {
	return &RemoveOutlier{
		sigma:        sigma,
		distanceFunc: HaversinDistance,
		metric:       "Speed",
		unit:         "m/s",
		value: func(line *line) *float64 {
			return line.speed
		},
	}
}

func RemoveOutlierByDistance(sigma int) *RemoveOutlier {
	return &RemoveOutlier{
		sigma:        sigma,
		distanceFunc: HaversinDistance,
		metric:       "Distance",
		unit:         "m",
		value: func(line *line) *float64 {
			return &line.dist
		},
	}
}

type RemoveOutlier struct {
	sigma        int
	distanceFunc DistanceFunc
	metric       string
	unit         string
	value        func(line *line) *float64
}

func (r *RemoveOutlier) Run(tracklog *gpx.TrackLog) (int, error) {
	n := 0
	for _, t := range tracklog.Tracks {
		for i, seg := range t.Segments {
			num := len(seg.Points)
			removed, err := r.remove(seg)
			if err != nil {
				return 0, err
			}
			numRemoved := (num - len(removed.Points))
			log.Debugf("Removed %d points", numRemoved)
			n += numRemoved
			t.Segments[i] = removed
		}
	}
	return n, nil
}

func (r *RemoveOutlier) remove(seg *gpx.Segment) (*gpx.Segment, error) {
	lines := getLines(r.distanceFunc, seg.Points)
	sum := 0.0
	num := 0
	for _, line := range lines {
		value := r.value(line)
		if value != nil {
			sum += *value
			num++
		}
	}
	avg := sum / float64(num)
	log.Debugf("Average: %f %s", avg, r.unit)

	std := 0.0
	for _, line := range lines {
		value := r.value(line)
		if value != nil {
			std += math.Pow(*value-avg, 2)
		}
	}
	std = math.Sqrt(std / float64(num))
	log.Debugf("Standard deviation: %f %s", std, r.unit)

	sigma := float64(r.sigma) * std // three sigma
	log.Debugf("%d-Sigma: %f %s", r.sigma, sigma, r.unit)

	accepted := make([]*line, 0)
	for _, line := range lines {
		value := r.value(line)
		if value != nil {
			// log.Debugf("Speed %v", *value)
			if math.Abs(*value)-avg > sigma {
				log.Debugf("Discarding %v %s", *value, r.unit)
				continue
			}
		}
		accepted = append(accepted, line)
	}

	return &gpx.Segment{
		Points: joinLines(accepted),
	}, nil
}
