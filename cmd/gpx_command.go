package cmd

import (
	"fmt"
	"gpxtoolkit/gpx"
	"io"
	"os"
	"strconv"
	"time"
)

type GPXCommand struct {
	Output io.Writer
}

func (c *GPXCommand) arg(args []string, index int) (string, error) {
	if len(args) <= index {
		return "", fmt.Errorf("Index of out bound: %d vs %v", index, args)
	}
	return args[index], nil
}

func (c *GPXCommand) floatArg(args []string, index int) (float64, error) {
	str, err := c.arg(args, index)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(str, 64)
}

func (c *GPXCommand) timeArg(args []string, index int) (time.Time, error) {
	var tm time.Time
	str, err := c.arg(args, index)
	if err != nil {
		return tm, err
	}
	return time.Parse("2006-01-02 15:04:05", str)
}

func (c *GPXCommand) openGpx(args []string, index int) (*gpx.TrackLog, error) {
	file, err := c.arg(args, index)
	if err != nil {
		return nil, err
	}
	gpxLog, err := gpx.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open GPX '%s': %s\n", file, err.Error())
		return nil, err
	}
	return gpxLog, nil
}

func (c *GPXCommand) writeGpx(gpxLog *gpx.TrackLog, creator string) error {
	writer := &gpx.Writer{
		Creator: creator,
		Writer:  c.Output,
	}
	err := writer.Write(gpxLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write GPX: %s\n", err.Error())
		return err
	}
	return nil
}
