package cmd

import (
	"fmt"
	"gpxtoolkit/gpx"
	"math"
	"os"
)

type CreateSVG struct {
	InputFile  string
	ZoomLevel  int
	TileWidth  int
	TileHeight int
	Padding    struct {
		Top     int
		Left    int
		Botttom int
		Right   int
	}
}

type tileImage struct {
	URL     string
	Opacity float32
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

func (c *CreateSVG) Run() error {
	log, err := gpx.Open(c.InputFile)
	if err != nil {
		return fmt.Errorf("Failed to open GPX '%s': %s", c.InputFile, err.Error())
	}
	bbox := log.BoundingBox()
	if bbox.Min == nil || bbox.Max == nil {
		return fmt.Errorf("Failed to calculate bounding box: %s", c.InputFile)
	}
	file := os.Stdout
	minX, minY := c.getIntXY(bbox.Min.Latitude, bbox.Min.Longitude)
	maxX, maxY := c.getIntXY(bbox.Max.Latitude, bbox.Max.Longitude)
	minX -= c.Padding.Left
	minY += c.Padding.Botttom
	maxX += c.Padding.Right
	maxY -= c.Padding.Top
	width := (maxX - minX + 1) * c.TileWidth
	height := (minY - maxY + 1) * c.TileHeight
	maxWidthHeight := math.Max(float64(width), float64(height))
	fmt.Fprintf(file, "<svg width=\"%d\" height=\"%d\" xmlns=\"http://www.w3.org/2000/svg\">\n", width, height)
	fmt.Fprintf(file, "  <rect width=\"%d\" height=\"%d\" fill=\"#80ff80\"/>\n", width, height)
	for x := minX; x <= maxX; x++ {
		for y := maxY; y <= minY; y++ {
			dx := (x - minX) * c.TileWidth
			dy := (y - maxY) * c.TileHeight
			images := c.getImages(x, y)
			for _, image := range images {
				fmt.Fprintf(file, "  <image href=\"%s\" x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" opacity=\"%f\"/>\n", image.URL, dx, dy, c.TileWidth, c.TileHeight, image.Opacity)
			}
		}
	}
	// floatMinX, _ := c.getXY(bbox.Min.Latitude, bbox.Min.Longitude)
	// _, floatMaxY := c.getXY(bbox.Max.Latitude, bbox.Max.Longitude)
	// shiftX := (floatMinX - float64(minX)) * float64(c.TileWidth)
	// shiftY := (float64(maxY) - floatMaxY) * float64(c.TileHeight)
	for _, t := range log.GetTracks() {
		for _, s := range t.GetSegments() {
			styles := []routeStyle{
				{Stroke: "#ffffff", Width: int(math.Max(1, maxWidthHeight*0.003)), Opacity: 0.5},
				{Stroke: "#0000ff", Width: int(math.Max(1, maxWidthHeight*0.001)), Opacity: 0.5},
			}
			for _, st := range styles {
				fmt.Fprintf(file, "  <polyline fill=\"none\" stroke=\"%s\" stroke-width=\"%d\" opacity=\"%f\" points=\"", st.Stroke, st.Width, st.Opacity)
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
		}
	}
	for _, p := range log.WayPoints {
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
			fmt.Fprintf(file, "  <circle cx=\"%f\" cy=\"%f\" r=\"%d\" fill=\"%s\" opacity=\"%f\"/>", dx, dy, s.Radius, s.Fill, s.Opacity)
		}
		textStyles := []textStyle{
			{Fill: "none", Stroke: "#ffffff", StrokeWidth: int(math.Max(1, maxWidthHeight*0.001)), Opacity: 0.9},
			{Fill: "#000000", Stroke: "none", Opacity: 1.0},
		}
		for _, st := range textStyles {
			fmt.Fprintf(file, "  <text x=\"%f\" y=\"%f\" fill=\"%s\" stroke=\"%s\" stroke-width=\"%d\" opacity=\"%f\">%s</text>", dx+shift, dy-shift, st.Fill, st.Stroke, st.StrokeWidth, st.Opacity, p.GetName())
		}
	}
	fmt.Fprintf(file, "</svg>")
	return nil
}

func (c *CreateSVG) getIntXY(lat, lon float64) (int, int) {
	x, y := c.getXY(lat, lon)
	return int(math.Floor(x)), int(math.Floor(y))
}

func (c *CreateSVG) getXY(lat, lon float64) (x, y float64) {
	x = (lon + 180.0) / 360.0 * (math.Exp2(float64(c.ZoomLevel)))
	y = (1.0 - math.Log(math.Tan(lat*math.Pi/180.0)+1.0/math.Cos(lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(c.ZoomLevel)))
	return
}

func (c *CreateSVG) getImages(x, y int) []*tileImage {
	images := make([]*tileImage, 0)
	// return fmt.Sprintf("https://a.tile.openstreetmap.org/%d/%d/%d.png", c.ZoomLevel, x, y)
	// images = append(images, &tileImage{
	// 	URL:     fmt.Sprintf("https://wmts.nlsc.gov.tw/wmts/EMAP98/default/EPSG:3857/%d/%d/%d.png", c.ZoomLevel, y, x),
	// 	Opacity: 1.0,
	// })
	// images = append(images, &tileImage{
	// 	URL:     fmt.Sprintf("https://wmts.nlsc.gov.tw/wmts/EMAP16/default/EPSG:3857/%d/%d/%d.png", c.ZoomLevel, y, x),
	// 	Opacity: 1.0,
	// })
	// images = append(images, &tileImage{
	// 	URL:     fmt.Sprintf("https://wmts.nlsc.gov.tw/wmts/EMAP6/default/EPSG:3857/%d/%d/%d.png", c.ZoomLevel, y, x),
	// 	Opacity: 1.0,
	// })
	images = append(images, &tileImage{
		URL:     fmt.Sprintf("https://wmts.nlsc.gov.tw/wmts/MOI_HILLSHADE/default/EPSG:3857/%d/%d/%d.png", c.ZoomLevel, y, x),
		Opacity: 0.6,
	})
	images = append(images, &tileImage{
		URL:     fmt.Sprintf("https://wmts.nlsc.gov.tw/wmts/MOI_CONTOUR_2/default/EPSG:3857/%d/%d/%d.png", c.ZoomLevel, y, x),
		Opacity: 1.0,
	})
	images = append(images, &tileImage{
		URL:     fmt.Sprintf("https://wmts.nlsc.gov.tw/wmts/EMAP2/default/EPSG:3857/%d/%d/%d.png", c.ZoomLevel, y, x),
		Opacity: 1.0,
	})
	return images
}
