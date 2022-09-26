package gpx

import (
	"math"
	"os"
	"testing"
	"time"
)

func TestStats(t *testing.T) {
	path := "tests/hiking_d2254ab62217fe37d259f2052f31b74a.gpx"
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p := &Parser{}
	log, err := p.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	st, err := log.Stat(1.0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Round(*st.ElevationGain) != 464 {
		t.Fatalf("Unexpected elevation gain: %v", *st.ElevationGain)
	}
	if math.Round(*st.ElevationLoss) != 533 {
		t.Fatalf("Unexpected elevation loss: %v", *st.ElevationLoss)
	}
	if math.Round(*st.Distance) != 3843 {
		t.Fatalf("Unexpected distance: %v", *st.Distance)
	}
	du, _ := time.ParseDuration("3h30m4s")
	if st.Duration() != du {
		t.Fatalf("Unexpected duration: %v", st.Duration())
	}
}
