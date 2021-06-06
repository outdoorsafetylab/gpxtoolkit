package gpx

import (
	"gpxtoolkit/xml"
	"io"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"
)

type Parser struct {
}

func (p *Parser) Parse(r io.Reader) (*TrackLog, error) {
	var log *TrackLog
	var track *Track
	var segment *Segment
	var pt *Point
	var wpt *WayPoint
	err := xml.NewParser().On("//gpx", func(attrs map[string]string) error {
		log = &TrackLog{Tracks: make([]*Track, 0)}
		creator := attrs["creator"]
		if creator != "" {
			log.Creator = proto.String(creator)
		}
		return nil
	}, nil, nil).OnText("//gpx/metadata/name", true, func(text string) error {
		log.Name = proto.String(text)
		return nil
	}).OnText("//gpx/metadata/time", true, func(text string) error {
		tm, err := time.Parse(time.RFC3339, text)
		if err != nil {
			return err
		}
		log.NanoTime = proto.Int64(tm.UnixNano())
		return nil
	}).On("//gpx/metadata/link", func(attrs map[string]string) error {
		link := attrs["href"]
		if link != "" {
			log.Link = &TrackLink{Url: proto.String(link)}
		}
		return nil
	}, nil, nil).OnText("//gpx/metadata/link/text", true, func(text string) error {
		if log.Link != nil {
			log.Link.Text = proto.String(text)
		}
		return nil
	}).On("//gpx/trk", func(map[string]string) error {
		track = &Track{Segments: make([]*Segment, 0)}
		return nil
	}, nil, func() error {
		log.Tracks = append(log.Tracks, track)
		track = nil
		return nil
	}).OnText("//gpx/trk/name", true, func(text string) error {
		track.Name = proto.String(text)
		return nil
	}).OnText("//gpx/trk/type", true, func(text string) error {
		track.Type = proto.String(text)
		return nil
	}).OnText("//gpx/trk/cmt", true, func(text string) error {
		track.Comment = proto.String(text)
		return nil
	}).On("//gpx/trk/trkseg", func(map[string]string) error {
		segment = &Segment{Points: make([]*Point, 0)}
		return nil
	}, nil, func() error {
		track.Segments = append(track.Segments, segment)
		segment = nil
		return nil
	}).On("//gpx/trk/trkseg/trkpt", func(attrs map[string]string) (err error) {
		pt = &Point{}
		lat, err := strconv.ParseFloat(attrs["lat"], 64)
		if err != nil {
			return err
		}
		pt.Latitude = proto.Float64(lat)
		lon, err := strconv.ParseFloat(attrs["lon"], 64)
		if err != nil {
			return err
		}
		pt.Longitude = proto.Float64(lon)
		return nil
	}, nil, func() error {
		segment.Points = append(segment.Points, pt)
		pt = nil
		return nil
	}).OnText("//gpx/trk/trkseg/trkpt/ele", true, func(text string) error {
		elev, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return err
		}
		pt.Elevation = new(float64)
		*pt.Elevation = elev
		return nil
	}).OnText("//gpx/trk/trkseg/trkpt/time", true, func(text string) error {
		tm, err := time.Parse(time.RFC3339, text)
		if err != nil {
			return err
		}
		pt.NanoTime = proto.Int64(tm.UnixNano())
		return nil
	}).On("//gpx/wpt", func(attrs map[string]string) error {
		wpt = &WayPoint{}
		lat, err := strconv.ParseFloat(attrs["lat"], 64)
		if err != nil {
			return err
		}
		wpt.Latitude = proto.Float64(lat)
		lon, err := strconv.ParseFloat(attrs["lon"], 64)
		if err != nil {
			return err
		}
		wpt.Longitude = proto.Float64(lon)
		return nil
	}, nil, func() error {
		log.WayPoints = append(log.WayPoints, wpt)
		wpt = nil
		return nil
	}).OnText("//gpx/wpt/ele", true, func(text string) error {
		elev, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return err
		}
		wpt.Elevation = new(float64)
		*wpt.Elevation = elev
		return nil
	}).OnText("//gpx/wpt/time", true, func(text string) error {
		tm, err := time.Parse(time.RFC3339, text)
		if err != nil {
			return err
		}
		wpt.NanoTime = proto.Int64(tm.UnixNano())
		return nil
	}).OnText("//gpx/wpt/name", true, func(text string) (err error) {
		wpt.Name = proto.String(text)
		return err
	}).OnText("//gpx/wpt/desc", true, func(text string) (err error) {
		wpt.Description = proto.String(text)
		return err
	}).OnText("//gpx/wpt/cmt", true, func(text string) (err error) {
		wpt.Comment = proto.String(text)
		return err
	}).OnText("//gpx/wpt/sym", true, func(text string) (err error) {
		wpt.Symbol = proto.String(text)
		return err
	}).Parse(r)
	if err != nil {
		return nil, err
	}
	return log, nil
}
