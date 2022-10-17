package svg

import (
	"encoding/base64"
	"fmt"
	"gpxtoolkit/gpx"
	"io"
	"math"
	"net/http"
	"os"
)

type Create struct {
	ZoomLevel   int
	TileWidth   int
	TileHeight  int
	Background  string
	TilePadding struct {
		Top     int
		Left    int
		Botttom int
		Right   int
	}
	EmbedImage bool
	Scale      struct {
		Unit       int
		Repeat     int
		Stroke     string
		FillColors []string
	}
}

type layer struct {
	Name          string
	TileURLFormat string
	Opacity       float32
}

type routeStyle struct {
	Stroke  string
	Width   int
	Opacity float32
}

type markerStyle struct {
	Fill    string
	Radius  int
	Opacity float32
}

type textStyle struct {
	Fill        string
	Stroke      string
	StrokeWidth int
	Opacity     float32
}

func (c *Create) Run(log *gpx.TrackLog) error {
	bbox := log.BoundingBox()
	if bbox.Min == nil || bbox.Max == nil {
		return fmt.Errorf("Failed to calculate bounding box")
	}
	file := os.Stdout
	minX, minY := c.getIntXY(bbox.Min.Latitude, bbox.Min.Longitude)
	maxX, maxY := c.getIntXY(bbox.Max.Latitude, bbox.Max.Longitude)
	minX -= c.TilePadding.Left
	minY += c.TilePadding.Botttom
	maxX += c.TilePadding.Right
	maxY -= c.TilePadding.Top
	width := (maxX - minX + 1) * c.TileWidth
	height := (minY - maxY + 1) * c.TileHeight
	maxWidthHeight := math.Max(float64(width), float64(height))
	fmt.Fprintf(file, "<svg width=\"%d\" height=\"%d\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\">\n", width, height)
	fmt.Fprintf(file, "  <g id=\"background\">\n")
	fmt.Fprintf(file, "    <rect width=\"%d\" height=\"%d\" fill=\"%s\"/>\n", width, height, c.Background)
	fmt.Fprintf(file, "  </g>\n")
	layers := []layer{
		{Name: "hillshading", Opacity: 0.6, TileURLFormat: "https://wmts.nlsc.gov.tw/wmts/MOI_HILLSHADE/default/EPSG:3857/%d/%d/%d.png"},
		{Name: "contour", Opacity: 0.5, TileURLFormat: "https://wmts.nlsc.gov.tw/wmts/MOI_CONTOUR_2/default/EPSG:3857/%d/%d/%d.png"},
		{Name: "water", Opacity: 0.5, TileURLFormat: "https://wmts.nlsc.gov.tw/wmts/LUIMAP04/default/EPSG:3857/%d/%d/%d.png"},
		{Name: "map", Opacity: 1.0, TileURLFormat: "https://wmts.nlsc.gov.tw/wmts/EMAP2/default/EPSG:3857/%d/%d/%d.png"},
	}
	for _, l := range layers {
		fmt.Fprintf(file, "  <g id=\"%s\" opacity=\"%f\">\n", l.Name, l.Opacity)
		for x := minX; x <= maxX; x++ {
			for y := maxY; y <= minY; y++ {
				dx := (x - minX) * c.TileWidth
				dy := (y - maxY) * c.TileHeight
				fmt.Fprintf(file, "    <image x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" ", dx, dy, c.TileWidth, c.TileHeight)
				imageUrl := l.getURL(x, y, c.ZoomLevel)
				if c.EmbedImage {
					fmt.Fprintf(os.Stderr, "Embedding: %s\n", imageUrl)
					res, err := http.Get(imageUrl)
					if err != nil {
						return err
					}
					defer res.Body.Close()
					if res.StatusCode != 200 {
						return fmt.Errorf("Failed to download image '%s': %s", imageUrl, res.Status)
					}
					fmt.Fprintf(file, "xlink:href=\"data:image/png;base64,")
					err = c.base64enc(res.Body, file)
					if err != nil {
						return err
					}
					fmt.Fprintf(file, "\"")
				} else {
					fmt.Fprintf(file, "href=\"%s\"", imageUrl)
				}
				fmt.Fprintf(file, " />\n")
			}
		}
		fmt.Fprintf(file, "  </g>\n")
	}
	fmt.Fprintf(file, "  <g id=\"scale\">\n")
	meterPerPixel := 40075016.686 * math.Cos((bbox.Max.Latitude-bbox.Min.Latitude)/2*(math.Pi/180)) / math.Pow(2, float64(c.ZoomLevel)) / float64(c.TileWidth)
	w := float64(c.Scale.Unit) / meterPerPixel
	h := math.Max(1, maxWidthHeight*0.005)
	for i := 0; i < c.Scale.Repeat; i++ {
		x := 50.0 + float64(i)*w
		y := float64(height) - 50 - h
		sw := int(math.Max(1, maxWidthHeight*0.001))
		fill := c.Scale.FillColors[i%len(c.Scale.FillColors)]
		fmt.Fprintf(file, "    <rect x=\"%f\" y=\"%f\" width=\"%f\" height=\"%f\" stroke=\"%s\" stroke-width=\"%d\" fill=\"%s\"/>\n", x, y, w, h, c.Scale.Stroke, sw, fill)
	}
	fmt.Fprintf(file, "  </g>\n")
	for i, t := range log.GetTracks() {
		fmt.Fprintf(file, "  <g id=\"track_%02d_%s\">\n", i, t.GetName())
		for j, s := range t.GetSegments() {
			fmt.Fprintf(file, "    <g id=\"track_%02d_segment_%02d\">\n", i, j)
			styles := []routeStyle{
				{Stroke: "#ffffff", Width: int(math.Max(1, maxWidthHeight*0.003)), Opacity: 0.5},
				{Stroke: "#0000ff", Width: int(math.Max(1, maxWidthHeight*0.001)), Opacity: 0.5},
			}
			for _, st := range styles {
				fmt.Fprintf(file, "      <polyline fill=\"none\" stroke=\"%s\" stroke-width=\"%d\" opacity=\"%f\" points=\"", st.Stroke, st.Width, st.Opacity)
				for _, p := range s.GetPoints() {
					if p.Latitude == nil || p.Longitude == nil {
						continue
					}
					x, y := c.getXY(p.GetLatitude(), p.GetLongitude())
					dx := (x - float64(minX)) * float64(c.TileWidth)
					dy := (y - float64(maxY)) * float64(c.TileHeight)
					fmt.Fprintf(file, "%f,%f ", dx, dy)
				}
				fmt.Fprintf(file, "\"/>\n")
			}
			fmt.Fprintf(file, "    </g>\n")
		}
		fmt.Fprintf(file, "  </g>\n")
	}
	for i, p := range log.WayPoints {
		fmt.Fprintf(file, "  <g id=\"waypoint_%02d_%s\">\n", i, p.GetName())
		if p.Latitude == nil || p.Longitude == nil {
			continue
		}
		x, y := c.getXY(p.GetLatitude(), p.GetLongitude())
		dx := (x - float64(minX)) * float64(c.TileWidth)
		dy := (y - float64(maxY)) * float64(c.TileHeight)
		styles := []markerStyle{
			{Fill: "#ffffff", Radius: int(math.Max(1, maxWidthHeight*0.003)), Opacity: 0.5},
			{Fill: "#ff0000", Radius: int(math.Max(1, maxWidthHeight*0.002)), Opacity: 0.8},
		}
		shift := 0.0
		for _, s := range styles {
			shift = math.Max(shift, float64(s.Radius))
			fmt.Fprintf(file, "    <circle cx=\"%f\" cy=\"%f\" r=\"%d\" fill=\"%s\" opacity=\"%f\"/>\n", dx, dy, s.Radius, s.Fill, s.Opacity)
		}
		textStyles := []textStyle{
			{Fill: "none", Stroke: "#ffffff", StrokeWidth: int(math.Max(1, maxWidthHeight*0.001)), Opacity: 0.9},
			{Fill: "#000000", Stroke: "none", Opacity: 1.0},
		}
		fontSize := int(math.Max(1, maxWidthHeight*0.008))
		for _, st := range textStyles {
			fmt.Fprintf(file, "    <text x=\"%f\" y=\"%f\" font-size=\"%d\" fill=\"%s\" stroke=\"%s\" stroke-width=\"%d\" opacity=\"%f\">%s</text>\n", dx+shift, dy-shift, fontSize, st.Fill, st.Stroke, st.StrokeWidth, st.Opacity, p.GetName())
		}
		fmt.Fprintf(file, "  </g>\n")
	}
	fmt.Fprintf(file, "</svg>")
	return nil
}

func (c *Create) getIntXY(lat, lon float64) (int, int) {
	x, y := c.getXY(lat, lon)
	return int(math.Floor(x)), int(math.Floor(y))
}

func (c *Create) getXY(lat, lon float64) (x, y float64) {
	x = (lon + 180.0) / 360.0 * (math.Exp2(float64(c.ZoomLevel)))
	y = (1.0 - math.Log(math.Tan(lat*math.Pi/180.0)+1.0/math.Cos(lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(c.ZoomLevel)))
	return
}

func (c *Create) base64enc(r io.Reader, w io.Writer) error {
	pr, pw := io.Pipe()
	encoder := base64.NewEncoder(base64.StdEncoding, pw)
	go func() {
		_, err := io.Copy(encoder, r)
		encoder.Close()
		if err != nil {
			_ = pw.CloseWithError(err)
		} else {
			pw.Close()
		}
	}()
	_, err := io.Copy(w, pr)
	if err != nil {
		_ = pr.CloseWithError(err)
	} else {
		pr.Close()
	}
	return err
}

func (l layer) getURL(x, y, z int) string {
	return fmt.Sprintf(l.TileURLFormat, z, y, x)
}
