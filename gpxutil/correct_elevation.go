package gpxutil

import (
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"log"
	"math"

	"google.golang.org/protobuf/proto"
)

type CorrectElevation struct {
	Waypoints bool
	Service   elevation.Service
}

func (c *CorrectElevation) Name() string {
	return "Correct Elevation"
}

func (c *CorrectElevation) Run(tracklog *gpx.TrackLog) (int, error) {
	if c.Service == nil {
		return 0, nil
	}
	n := 0
	if c.Waypoints {
		m, err := correctWayPoints(c.Service, tracklog.WayPoints)
		if err != nil {
			return 0, err
		}
		n += m
	}
	for _, t := range tracklog.Tracks {
		for _, s := range t.Segments {
			m, err := correctPoints(c.Service, s.Points)
			if err != nil {
				return 0, err
			}
			n += m
		}
	}
	return n, nil
}

func correctWayPoints(service elevation.Service, waypoints []*gpx.WayPoint) (int, error) {
	points := make([]*elevation.LatLon, 0)
	for _, p := range waypoints {
		points = append(points, &elevation.LatLon{Lat: p.GetLatitude(), Lon: p.GetLongitude()})
	}
	elevations, err := service.Lookup(points)
	if err != nil {
		return 0, err
	}
	n := 0
	for i, p := range waypoints {
		elev := elevations[i]
		if elev == nil || math.IsNaN(*elev) {
			continue
		}
		p.Elevation = proto.Float64(math.Round(*elev))
		n++
	}
	log.Printf("Corrected %d way points", n)
	return n, nil
}

func correctPoints(service elevation.Service, gpxPoints []*gpx.Point) (int, error) {
	points := make([]*elevation.LatLon, 0)
	for _, p := range gpxPoints {
		points = append(points, &elevation.LatLon{Lat: p.GetLatitude(), Lon: p.GetLongitude()})
	}
	elevations, err := service.Lookup(points)
	if err != nil {
		return 0, err
	}
	n := 0
	for i, p := range gpxPoints {
		elev := elevations[i]
		if elev == nil || math.IsNaN(*elev) {
			continue
		}
		p.Elevation = proto.Float64(math.Round(*elev))
		n++
	}
	log.Printf("Corrected %d points", n)
	return n, nil
}
