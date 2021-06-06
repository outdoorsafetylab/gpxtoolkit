package gpx

import (
	"io"
	"os"
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
