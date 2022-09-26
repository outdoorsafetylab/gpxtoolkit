package gpxutil

import (
	"fmt"
	"math"

	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/log"

	"github.com/maja42/goval"
	"google.golang.org/protobuf/proto"
)

type Milestone struct {
	Service           elevation.Service
	Distance          float64
	MilestoneName     *MilestoneName
	Symbol            string
	Reverse           bool
	FitWaypoints      bool
	ByTerrainDistance bool
	distanceFunc      DistanceFunc
}

func (c *Milestone) Name() string {
	return "Create Milestones"
}

func (c *Milestone) Run(tracklog *gpx.TrackLog) (int, error) {
	if c.ByTerrainDistance {
		c.distanceFunc = TerrainDistance
	} else {
		c.distanceFunc = HaversinDistance
	}
	n := 0
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			points := seg.Points
			if c.Reverse {
				points = make([]*gpx.Point, len(seg.Points))
				for i, p := range seg.Points {
					points[len(seg.Points)-1-i] = p
				}
			}
			if c.ByTerrainDistance && c.Service != nil {
				// we don't alter the original points
				corrected := make([]*gpx.Point, len(points))
				copy(corrected, points)
				_, err := correctPoints(c.Service, corrected)
				if err != nil {
					return 0, err
				}
				points = corrected
			}
			waypoints := tracklog.WayPoints
			if !c.FitWaypoints {
				waypoints = nil
			}
			milestones, err := c.milestone(points, waypoints)
			if err != nil {
				return 0, err
			}
			n += len(milestones)
			log.Debugf("Appending %d milestones", len(milestones))
			tracklog.WayPoints = append(tracklog.WayPoints, milestones...)
		}
	}
	return n, nil
}

type milestone struct {
	distance  float64
	variables *MilestoneNameVariables
	waypoint  *gpx.WayPoint
}

func (c *Milestone) milestone(points []*gpx.Point, waypoints []*gpx.WayPoint) ([]*gpx.WayPoint, error) {
	if len(points) <= 0 {
		return []*gpx.WayPoint{}, nil
	}
	if waypoints == nil {
		distances := make([]float64, len(points)-1)
		total := 0.0
		for i, b := range points[1:] {
			a := points[i]
			dist := c.distanceFunc(a, b)
			distances[i] = dist
			total += dist
		}
		milestones := make([]*milestone, int(math.Floor(total/c.Distance)))
		for i := range milestones {
			milestones[i] = &milestone{
				variables: &MilestoneNameVariables{
					Number:   i + 1,
					Total:    len(milestones),
					Distance: float64(i+1) * c.Distance,
				},
				distance: float64(i+1) * c.Distance,
			}
		}
		log.Debugf("Total %d points: %.1fm with %d milestones", len(points), total, len(milestones))
		return c.create(points, milestones, distances)
	} else {
		segments, err := sliceByWaypoints(c.distanceFunc, points, waypoints, c.Distance/2)
		if err != nil {
			return nil, err
		}
		log.Debugf("Sliced %d points to %d segments", len(points), len(segments))
		lengths := make([]float64, len(segments))
		distances := make([][]float64, len(segments))
		numMilestones := make([]int, len(segments))
		totalMilestones := 0
		numPoints := 0
		distance := 0.0
		for i, segment := range segments {
			log.Debugf("Segment %d: %d points", i, len(segment.points))
			numPoints += len(segment.points)
			start := distance
			distances[i] = make([]float64, len(segment.points)-1)
			for j, b := range segment.points[1:] {
				a := segment.points[j]
				dist := c.distanceFunc(a, b)
				distances[i][j] = dist
				distance += dist
			}
			length := (distance - start)
			numMilestone := int(math.Round(length / c.Distance))
			a := "start"
			if segment.a.waypoint != nil {
				a = segment.a.waypoint.GetName()
			}
			b := "end"
			if segment.b.waypoint != nil {
				b = segment.b.waypoint.GetName()
			}
			log.Debugf("Segment %d: from %s @ %.1fm to %s @ %.1fm: %f meters with %d milestones", i, a, start, b, distance, length, numMilestone)
			lengths[i] = length
			numMilestones[i] = numMilestone
			totalMilestones += numMilestone
		}
		log.Debugf("Total %d points: %.1fm with %d milestones", numPoints, distance, totalMilestones)
		markers := make([]*gpx.WayPoint, 0)
		n := 0
		for i, segment := range segments {
			milestones := make([]*milestone, numMilestones[i])
			distance := lengths[i] / float64(len(milestones))
			for j := range milestones {
				milestones[j] = &milestone{
					variables: &MilestoneNameVariables{
						Number:   n + 1,
						Total:    totalMilestones,
						Distance: float64(n+1) * c.Distance,
					},
					distance: float64(j+1) * distance,
				}
				n++
			}
			if len(milestones) > 0 {
				milestones[len(milestones)-1].waypoint = segment.b.waypoint
			}
			m, err := c.create(segment.points, milestones, distances[i])
			if err != nil {
				return nil, err
			}
			markers = append(markers, m...)
		}
		return markers, nil
	}
}

func (c *Milestone) create(points []*gpx.Point, milestones []*milestone, distances []float64) ([]*gpx.WayPoint, error) {
	log.Debugf("Creating %d milestones from %d points ", len(milestones), len(points))
	markers := make([]*gpx.WayPoint, 0)
	start := 0.0
	for i, b := range points[1:] {
		a := points[i]
		var dist float64
		if distances != nil {
			dist = distances[i]
		} else {
			dist = c.distanceFunc(a, b)
		}
		end := start + dist
		// log.Printf("Current distance: %f", end)
		for _, ms := range milestones {
			// log.Printf("Milestone %d: %s @ %f", j, ms.name, ms.distance)
			if int(start*1000) >= int(ms.distance*1000) || int(end*1000) < int(ms.distance*1000) {
				continue
			}
			// log.Printf("Hit milestone %s: %f => %v", ms.name, ms.distance, ms.waypoint)
			if ms.waypoint != nil {
				ms.variables.Latitude = ms.waypoint.GetLatitude()
				ms.variables.Longitude = ms.waypoint.GetLongitude()
				ms.variables.Elevation = ms.waypoint.GetElevation()
				if c.Service != nil {
					err := ms.variables.correctElevation(c.Service)
					if err != nil {
						return nil, err
					}
				}
				name, err := c.MilestoneName.Eval(ms.variables)
				if err != nil {
					return nil, err
				}
				if ms.waypoint.Name != nil {
					ms.waypoint.Name = proto.String(fmt.Sprintf("%s/%s", ms.waypoint.GetName(), name))
				} else {
					ms.waypoint.Name = proto.String(name)
				}
			} else {
				p := interpolate(a, b, (ms.distance-start)/dist)
				if c.Service != nil {
					_, err := correctPoints(c.Service, []*gpx.Point{p})
					if err != nil {
						return nil, err
					}
				}
				ms.variables.Latitude = p.GetLatitude()
				ms.variables.Longitude = p.GetLongitude()
				ms.variables.Elevation = p.GetElevation()
				name, err := c.MilestoneName.Eval(ms.variables)
				if err != nil {
					return nil, err
				}
				wpt := &gpx.WayPoint{
					Name:      proto.String(name),
					Latitude:  p.Latitude,
					Longitude: p.Longitude,
					NanoTime:  p.NanoTime,
					Elevation: p.Elevation,
				}
				if c.Symbol != "" {
					wpt.Symbol = proto.String(c.Symbol)
				}
				markers = append(markers, wpt)
			}
		}
		start += dist
	}
	return markers, nil
}

type MilestoneNameVariables struct {
	Number, Total       int
	Distance            float64
	Latitude, Longitude float64
	Elevation           float64
}

func (v *MilestoneNameVariables) correctElevation(service elevation.Service) error {
	elev, err := service.Lookup([]*elevation.LatLon{{Lat: v.Latitude, Lon: v.Longitude}})
	if err != nil {
		return err
	}
	if elevation.IsValid(elev[0]) {
		v.Elevation = *elev[0]
	}
	return nil
}

type MilestoneName struct {
	Template string
}

func (n *MilestoneName) Eval(variables *MilestoneNameVariables) (string, error) {
	vars := map[string]interface{}{
		"num":   variables.Number,
		"total": variables.Total,
		"dist":  variables.Distance,
		"lat":   variables.Latitude,
		"lon":   variables.Longitude,
		"elev":  variables.Elevation,
	}
	val, err := goval.NewEvaluator().Evaluate(n.Template, vars, functions)
	if err != nil {
		return "", err
	}
	str := fmt.Sprintf("%v", val)
	return str, err
}

var functions map[string]func(args ...interface{}) (interface{}, error) = map[string]func(args ...interface{}) (interface{}, error){
	"printf": func(args ...interface{}) (interface{}, error) {
		str := fmt.Sprintf("%s", args[0])
		return fmt.Sprintf(str, args[1:]...), nil
	},
	"round": func(args ...interface{}) (interface{}, error) {
		return mathFunc(math.Round, args...)
	},
	"floor": func(args ...interface{}) (interface{}, error) {
		return mathFunc(math.Floor, args...)
	},
	"ceil": func(args ...interface{}) (interface{}, error) {
		return mathFunc(math.Ceil, args...)
	},
}

func mathFunc(call func(float64) float64, args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("unexpected number of arguments")
	}
	switch v := args[0].(type) {
	case int:
		return int(call(float64(v))), nil
	case int8:
		return int(call(float64(v))), nil
	case int16:
		return int(call(float64(v))), nil
	case int32:
		return int(call(float64(v))), nil
	case int64:
		return int(call(float64(v))), nil
	case float32:
		return int(call(float64(v))), nil
	case float64:
		return int(call(float64(v))), nil
	default:
		return nil, fmt.Errorf("unexpected type: %v", args[0])
	}
}
