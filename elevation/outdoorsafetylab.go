package elevation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gpxtoolkit/log"
)

type OutdoorSafetyLab struct {
	Client *http.Client
	URL    string
	Token  string
}

func (s *OutdoorSafetyLab) Lookup(latLons []*LatLon) ([]*float64, error) {
	start := time.Now()
	defer func() {
		log.Debugf("Looked up %d points in %v", len(latLons), time.Since(start))
	}()
	points := make([][]float64, len(latLons))
	for i, latlon := range latLons {
		points[i] = []float64{latlon.Lon, latlon.Lat}
	}
	return s.lookup(points)
}

func (s *OutdoorSafetyLab) lookup(points [][]float64) ([]*float64, error) {
	data, err := json.Marshal(points)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/elevations", s.URL), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Header.Set("Content-Type", "application/json")
	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	alts := make([]*float64, 0)
	dec := json.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(&alts)
	if err != nil {
		log.Errorf("Failed to decode elevations: %s", err.Error())
		return nil, err
	}
	if len(alts) != len(points) {
		err = fmt.Errorf("unepxected number of result: expect=%d, was=%d", len(points), len(alts))
		log.Errorf("%s", err.Error())
		return nil, err
	}
	return alts, nil
}
