package gpxutil

import (
	"gpxtoolkit/gpx"

	"github.com/oskanberg/eif-go"
)

type RemoveOutlierByEIF struct {
	Threshold float64
}

func (r *RemoveOutlierByEIF) Name() string {
	return "Remove Outlier by EIF"
}

func (r *RemoveOutlierByEIF) Run(tracklog *gpx.TrackLog) (int, error) {
	n := 0
	for _, t := range tracklog.Tracks {
		for i, seg := range t.Segments {
			num := len(seg.Points)
			seg, err := r.remove(seg)
			if err != nil {
				return 0, err
			}
			n += (num - len(seg.Points))
			t.Segments[i] = seg
		}
	}
	return n, nil
}

func (r *RemoveOutlierByEIF) remove(seg *gpx.Segment) (*gpx.Segment, error) {
	points := make([][]float64, len(seg.Points))
	for i, p := range seg.Points {
		points[i] = []float64{
			p.GetLatitude(),
			p.GetLongitude(),
			// p.GetElevation(),
		}
	}
	res := &gpx.Segment{
		Points: make([]*gpx.Point, 0),
	}
	f := eif.NewForest(points, eif.WithMaxTreeDepth(12), eif.WithTrees(100))
	for i, p := range points {
		score := f.Score(p)
		if score > r.Threshold {
			// log.Printf("Discarding %v: score=%f", p, score)
			continue
		}
		res.Points = append(res.Points, seg.Points[i])
	}
	return res, nil
}
