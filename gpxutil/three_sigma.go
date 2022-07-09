package gpxutil

import (
	"fmt"
	"gpxtoolkit/gpx"
	"log"
	"math"
)

func (c *RemoveOutlier) Name() string {
	return fmt.Sprintf("Remove Outliers by %s", c.metric)
}

func RemoveOutlierBySpeed() *RemoveOutlier {
	return &RemoveOutlier{
		distanceFunc: HorizontalDistance,
		metric:       "Speed",
		unit:         "m/s",
		value: func(line *line) *float64 {
			return line.speed
		},
	}
}

func RemoveOutlierByDistance() *RemoveOutlier {
	return &RemoveOutlier{
		distanceFunc: HorizontalDistance,
		metric:       "Distance",
		unit:         "m",
		value: func(line *line) *float64 {
			return &line.dist
		},
	}
}

type RemoveOutlier struct {
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
			n += (num - len(removed.Points))
			t.Segments[i] = removed
		}
	}
	return n, nil
}

func (r *RemoveOutlier) remove(seg *gpx.Segment) (*gpx.Segment, error) {
	lines := getLines(r.distanceFunc, seg.Points)
	var sum, avg, std float64
	num := 0
	for _, line := range lines {
		value := r.value(line)
		if value != nil {
			sum += *value
			num++
		}
	}
	avg = sum / float64(num)
	log.Printf("Average: %f %s", avg, r.unit)

	for _, line := range lines {
		value := r.value(line)
		if value != nil {
			std += math.Pow(*value-avg, 2)
		}
	}
	std = math.Sqrt(std / float64(num))
	log.Printf("Standard deviation: %f %s", std, r.unit)

	std3 := 3 * std // three sigma
	log.Printf("3-Sigma: %f %s", std3, r.unit)

	accepted := make([]*line, 0)
	for _, line := range lines {
		value := r.value(line)
		if value != nil && math.Abs(*value-avg) > std3 {
			log.Printf("Discarding %v %s", *value, r.unit)
			continue
		}
		accepted = append(accepted, line)
	}

	return &gpx.Segment{
		Points: joinLines(accepted),
	}, nil
}
