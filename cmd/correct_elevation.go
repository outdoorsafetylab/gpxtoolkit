package cmd

import (
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpxutil"
	"gpxtoolkit/log"
	"io"
	"net/http"
	"net/url"
)

type CorrectElevation struct {
	GPXCommand
}

func (c *CorrectElevation) Usage(w io.Writer, progname, cmdname string) {
	fmt.Fprintf(w, "%s %s <gpxfile> <elevation_service_url> [elevation_service_token]\n", progname, cmdname)
}

func (c *CorrectElevation) MinArguments() int {
	return 2
}

func (c *CorrectElevation) Run(progname, cmdname string, args []string) error {
	gpxLog, err := c.openGpx(args, 0)
	if err != nil {
		return err
	}
	serviceUrl, err := c.arg(args, 1)
	if err != nil {
		return err
	}
	serviceToken := ""
	if len(args) > 2 {
		serviceToken, err = c.arg(args, 2)
		if err != nil {
			return err
		}
	}
	url, err := url.Parse(serviceUrl)
	if err != nil || (url.Scheme != "http" && url.Scheme != "https") {
		return fmt.Errorf("Invalid URL of elevation service: '%s'", serviceUrl)
	}
	log.Infof("Using elevation service: %s", serviceUrl)
	service := &elevation.OutdoorSafetyLab{
		Client: http.DefaultClient,
		URL:    serviceUrl,
		Token:  serviceToken,
	}
	_, err = (&gpxutil.CorrectElevation{
		Waypoints: true,
		Service:   service,
	}).Run(gpxLog)
	if err != nil {
		return err
	}
	err = c.writeGpx(gpxLog, progname)
	if err != nil {
		return err
	}
	return nil
}
