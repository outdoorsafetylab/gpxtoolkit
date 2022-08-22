package cmd

import (
	"fmt"
	"gpxtoolkit/gpxutil"
	"io"
)

type ProjectWaypoints struct {
	GPXCommand
}

func (c *ProjectWaypoints) Usage(w io.Writer, progname, cmdname string) {
	fmt.Fprintf(w, "%s %s <gpxfile> <threshold>\n", progname, cmdname)
}

func (c *ProjectWaypoints) MinArguments() int {
	return 2
}

func (c *ProjectWaypoints) Run(progname, cmdname string, args []string) error {
	gpxLog, err := c.openGpx(args, 0)
	if err != nil {
		return err
	}
	cmd := &gpxutil.ProjectWaypoints{
		DistanceFunc: gpxutil.TerrainDistance,
		KeepOriginal: true,
	}
	cmd.Threshold, err = c.floatArg(args, 1)
	if err != nil {
		return err
	}
	_, err = cmd.Run(gpxLog)
	if err != nil {
		return err
	}
	err = c.writeGpx(gpxLog, progname)
	if err != nil {
		return err
	}
	return nil
}
