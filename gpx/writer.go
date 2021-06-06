package gpx

import (
	"fmt"
	"io"
	"time"
)

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
