package cmd

import (
	"errors"
	"fmt"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"
	"io"

	"google.golang.org/protobuf/proto"
)

type SliceByWaypoints struct {
	GPXCommand
}

func (c *SliceByWaypoints) Usage(w io.Writer, progname, cmdname string) {
	fmt.Fprintf(w, "%s %s <gpxfile> <threshold>\n", progname, cmdname)
}

func (c *SliceByWaypoints) MinArguments() int {
	return 2
}

func (c *SliceByWaypoints) Run(progname, cmdname string, args []string) error {
	gpxLog, err := c.openGpx(args, 0)
	if err != nil {
		return err
	}
	switch len(gpxLog.Tracks) {
	case 0:
		return errors.New("no track in the gpx")
	case 1:
		break
	default:
		return errors.New("more than 1 track in the gpx")
	}
	switch len(gpxLog.Tracks[0].Segments) {
	case 0:
		return errors.New("no segment in the gpx")
	case 1:
		break
	default:
		return errors.New("more than 1 segment in the track")
	}
	threshold, err := c.floatArg(args, 1)
	if err != nil {
		return err
	}
	points := gpxLog.Tracks[0].Segments[0].Points
	slices, err := gpxutil.SliceByWaypoints(gpxutil.TerrainDistance, points, gpxLog.WayPoints, threshold)
	if err != nil {
		return err
	}
	gpxLog.Tracks = make([]*gpx.Track, len(slices))
	for i := range gpxLog.Tracks {
		slice := slices[i]
		track := &gpx.Track{
			Segments: []*gpx.Segment{
				{Points: slice.Points},
			},
		}
		if slice.Start != nil {
			if slice.End != nil {
				track.Name = proto.String(fmt.Sprintf("%s => %s", slice.Start.GetName(), slice.End.GetName()))
			} else {
				track.Name = proto.String(fmt.Sprintf("%s =>", slice.Start.GetName()))
			}
		} else if slice.End != nil {
			track.Name = proto.String(fmt.Sprintf("=> %s", slice.End.GetName()))
		}
		if slice.End != nil {
			slice.End.Name = proto.String(fmt.Sprintf("%s'", slice.End.GetName()))
			gpxLog.WayPoints = append(gpxLog.WayPoints, slice.End)
		}
		gpxLog.Tracks[i] = track
	}
	err = c.writeGpx(gpxLog, progname)
	if err != nil {
		return err
	}
	return nil
}
