package gpxutil

import (
	"gpxtoolkit/gpx"
	"gpxtoolkit/log"
)

type Command interface {
	Name() string
	Run(tracklog *gpx.TrackLog) (int, error)
}

type ChainedCommands struct {
	Commands []Command
}

func (c *ChainedCommands) Name() string {
	return "Chained Commands"
}

func (c *ChainedCommands) Run(tracklog *gpx.TrackLog) (int, error) {
	n := 0
	for _, c := range c.Commands {
		log.Debugf("Running: %s", c.Name())
		m, err := c.Run(tracklog)
		if err != nil {
			return n, err
		}
		log.Debugf("Processed %d points", m)
		n += m
	}
	return n, nil
}
