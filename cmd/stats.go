package cmd

import (
	"fmt"
	"io"
	"log"
)

type Statistics struct {
	GPXCommand
}

func (c *Statistics) Usage(w io.Writer, progname, cmdname string) {
	fmt.Fprintf(w, "%s %s <gpxfile> <speed in km/h> [start time in YYYY-MM-DD HH:mm:SS]\n", progname, cmdname)
}

func (c *Statistics) MinArguments() int {
	return 1
}

func (c *Statistics) Run(progname, cmdname string, args []string) error {
	gpxLog, err := c.openGpx(args, 0)
	if err != nil {
		return err
	}
	alpha := 1.0
	if err != nil {
		return err
	}
	if len(args) > 1 {
		alpha, err = c.floatArg(args, 1)
		if err != nil {
			return err
		}
		log.Printf("Alpha: %f", alpha)
	}
	stat := gpxLog.Stat(alpha)
	log.Printf("%v", stat)
	return nil
}
