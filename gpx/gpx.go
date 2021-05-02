package gpx

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"gpxtoolkit/xml"

	"google.golang.org/protobuf/proto"
)

func Open(file string) (*TrackLog, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	parser := &Parser{}
	return parser.Parse(r)
}

func Parse(r io.Reader) (*TrackLog, error) {
	return (&Parser{}).Parse(r)
}

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

type Writer struct {
	Creator string
	Writer  io.Writer
}

func (gw *Writer) Write(log *TrackLog) error {
	newline := "\n"
	indent := &indent{
		value: "  ",
	}
	w := gw.Writer
	if _, err := w.Write([]byte(fmt.Sprintf(`%s<?xml version="1.0" encoding="UTF-8"?>%s`, indent, newline))); err != nil {
		return err
	}
	if _, err := w.Write([]byte(fmt.Sprintf(`%s<gpx version="1.1" creator="%s" xmlns="http://www.topografix.com/GPX/1/1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd">%s`, indent, gw.Creator, newline))); err != nil {
		return err
	}
	indent.level++
	if log.Name != nil || log.NanoTime != nil || log.Link != nil {
		if _, err := w.Write([]byte(fmt.Sprintf(`%s<metadata>%s`, indent, newline))); err != nil {
			return err
		}
		indent.level++
		if log.Name != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<name><![CDATA[%s]]></name>%s`, indent, log.GetName(), newline))); err != nil {
				return err
			}
		}
		if log.Link != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<link href="%s"><text>%s</text></link>%s`, indent, log.Link.GetUrl(), log.Link.GetText(), newline))); err != nil {
				return err
			}
		}
		if log.NanoTime != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<time>%s</time>%s`, indent, log.Time().Format(time.RFC3339), newline))); err != nil {
				return err
			}
		}
		indent.level--
		if _, err := w.Write([]byte(fmt.Sprintf(`%s</metadata>%s`, indent, newline))); err != nil {
			return err
		}
	}
	for _, wpt := range log.WayPoints {
		if _, err := w.Write([]byte(fmt.Sprintf(`%s<wpt lat="%f" lon="%f">%s`, indent, wpt.GetLatitude(), wpt.GetLongitude(), newline))); err != nil {
			return err
		}
		indent.level++
		if wpt.Elevation != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<ele>%f</ele>%s`, indent, wpt.GetElevation(), newline))); err != nil {
				return err
			}
		}
		if wpt.NanoTime != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<time>%s</time>%s`, indent, wpt.Time().Format(time.RFC3339), newline))); err != nil {
				return err
			}
		}
		if wpt.Name != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<name><![CDATA[%s]]></name>%s`, indent, wpt.GetName(), newline))); err != nil {
				return err
			}
		}
		if wpt.Comment != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<cmt><![CDATA[%s]]></cmt>%s`, indent, wpt.GetComment(), newline))); err != nil {
				return err
			}
		}
		if wpt.Description != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<desc><![CDATA[%s]]></desc>%s`, indent, wpt.GetDescription(), newline))); err != nil {
				return err
			}
		}
		if wpt.Symbol != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<sym>%s</sym>%s`, indent, wpt.GetSymbol(), newline))); err != nil {
				return err
			}
		}
		indent.level--
		if _, err := w.Write([]byte(fmt.Sprintf(`%s</wpt>%s`, indent, newline))); err != nil {
			return err
		}
	}
	for _, track := range log.Tracks {
		if _, err := w.Write([]byte(fmt.Sprintf(`%s<trk>%s`, indent, newline))); err != nil {
			return err
		}
		indent.level++
		if track.Name != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<name><![CDATA[%s]]></name>%s`, indent, track.GetName(), newline))); err != nil {
				return err
			}
		}
		if track.Comment != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<cmt><![CDATA[%s]]></cmt>%s`, indent, track.GetComment(), newline))); err != nil {
				return err
			}
		}
		if track.Type != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<type>%s</type>%s`, indent, track.GetType(), newline))); err != nil {
				return err
			}
		}
		for _, segment := range track.Segments {
			if _, err := w.Write([]byte(fmt.Sprintf(`%s<trkseg>%s`, indent, newline))); err != nil {
				return err
			}
			indent.level++
			for _, pt := range segment.Points {
				if _, err := w.Write([]byte(fmt.Sprintf(`%s<trkpt lat="%f" lon="%f">%s`, indent, pt.GetLatitude(), pt.GetLongitude(), newline))); err != nil {
					return err
				}
				indent.level++
				if pt.Elevation != nil {
					if _, err := w.Write([]byte(fmt.Sprintf(`%s<ele>%f</ele>%s`, indent, pt.GetElevation(), newline))); err != nil {
						return err
					}
				}
				if pt.NanoTime != nil {
					if _, err := w.Write([]byte(fmt.Sprintf(`%s<time>%s</time>%s`, indent, pt.Time().Format(time.RFC3339), newline))); err != nil {
						return err
					}
				}
				indent.level--
				if _, err := w.Write([]byte(fmt.Sprintf(`%s</trkpt>%s`, indent, newline))); err != nil {
					return err
				}
			}
			indent.level--
			if _, err := w.Write([]byte(fmt.Sprintf(`%s</trkseg>%s`, indent, newline))); err != nil {
				return err
			}
		}
		indent.level--
		if _, err := w.Write([]byte(fmt.Sprintf(`%s</trk>%s`, indent, newline))); err != nil {
			return err
		}
	}
	indent.level--
	if _, err := w.Write([]byte(fmt.Sprintf(`%s</gpx>%s`, indent, newline))); err != nil {
		return err
	}
	return nil
}

type indent struct {
	value string
	level int
}

func (i *indent) String() string {
	res := ""
	for x := 0; x < i.level; x++ {
		res = res + i.value
	}
	return res
}
