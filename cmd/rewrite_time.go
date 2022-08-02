package cmd

import (
	"errors"
	"fmt"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"
	"io"
	"time"

	"google.golang.org/protobuf/proto"
)

type RewriteTime struct {
	GPXCommand
}

func (c *RewriteTime) Usage(w io.Writer, progname, cmdname string) {
	fmt.Fprintf(w, "%s %s <gpxfile> <speed in km/h> [start time in YYYY-MM-DD HH:mm:SS]\n", progname, cmdname)
}

func (c *RewriteTime) MinArguments() int {
	return 2
}

func (c *RewriteTime) Run(progname, cmdname string, args []string) error {
	gpxLog, err := c.openGpx(args, 0)
	if err != nil {
		return err
	}
	speed, err := c.floatArg(args, 1)
	if err != nil {
		return err
	}
	start := time.Now()
	if len(args) > 2 {
		start, err = c.timeArg(args, 2)
		if err != nil {
			return err
		}
	}
	for _, track := range gpxLog.Tracks {
		for _, segment := range track.Segments {
			err = c.rewrite(segment.Points, start, speed)
			if err != nil {
				return err
			}
			start = segment.Points[len(segment.Points)-1].Time()
		}
	}
	err = c.writeGpx(gpxLog, progname)
	if err != nil {
		return err
	}
	return nil
}

func (c *RewriteTime) rewrite(points []*gpx.Point, start time.Time, speed float64) error {
	if len(points) == 0 {
		return errors.New("no points")
	}
	points[0].NanoTime = proto.Int64(start.UnixNano())
	if len(points) == 1 {
		return nil
	}
	for i, b := range points[1:] {
		a := points[i]
		dist := gpxutil.TerrainDistance(a, b)
		duration := time.Second * time.Duration(dist/1000/speed*60*60)
		start = start.Add(duration)
		b.NanoTime = proto.Int64(start.UnixNano())
	}
	return nil
}
