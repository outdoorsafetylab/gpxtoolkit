package cmd

import (
	"crypto/sha1"
	"fmt"
	"gpxtoolkit/gpx"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	exifcommon "github.com/dsoprea/go-exif/v3/common"
	heicexif "github.com/dsoprea/go-heic-exif-extractor/v2"
	jpegexif "github.com/dsoprea/go-jpeg-image-structure/v2"
	riimage "github.com/dsoprea/go-utility/v2/image"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var (
	exifDirectory = "."
)

// exifCmd represents the exif command
var exifCmd = &cobra.Command{
	Use:   "exif",
	Args:  cobra.NoArgs,
	Short: "Extract EXIF to GPX format",
	RunE: func(cmd *cobra.Command, args []string) error {
		wpts, err := loadImageAsWaypoints()
		if err != nil {
			return err
		}
		trackLog := &gpx.TrackLog{
			WayPoints: wpts,
		}
		fmt.Fprintf(os.Stderr, "Read %d images\n", len(wpts))
		return dumpGpx(trackLog)
	},
}

func init() {
	rootCmd.AddCommand(exifCmd)
	exifCmd.Flags().StringVarP(&exifDirectory, "directory", "d", exifDirectory, "Folder for images")
}

func loadImageAsWaypoints() ([]*gpx.WayPoint, error) {
	waypoints := make([]*gpx.WayPoint, 0)
	heicParser := heicexif.NewHeicExifMediaParser()
	jpegParser := jpegexif.NewJpegMediaParser()
	list, err := ioutil.ReadDir(exifDirectory)
	if err != nil {
		return nil, err
	}
	for _, info := range list {
		if info.IsDir() {
			continue
		}
		file := filepath.Join(exifDirectory, info.Name())
		var mc riimage.MediaContext
		var err error
		if strings.HasSuffix(file, ".HEIC") || strings.HasSuffix(file, ".heic") {
			mc, err = heicParser.ParseFile(file)
		} else if strings.HasSuffix(file, ".JPEG") || strings.HasSuffix(file, ".jpeg") {
			mc, err = jpegParser.ParseFile(file)
		} else if strings.HasSuffix(file, ".JPG") || strings.HasSuffix(file, ".jpg") {
			mc, err = jpegParser.ParseFile(file)
		} else {
			continue
		}
		if err != nil {
			return nil, err
		}
		ifd, _, err := mc.Exif()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: Missing IFD/Exif\n", filepath.Base(file))
			return nil, err
		}
		gps, err := ifd.ChildWithIfdPath(exifcommon.IfdGpsInfoStandardIfdIdentity)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: Missing IFD/GPSInfo\n", filepath.Base(file))
			continue
		}
		info, err := gps.GpsInfo()
		if err != nil {
			return nil, err
		}
		lat := info.Latitude.Decimal()
		lng := info.Longitude.Decimal()
		alt := info.Altitude
		tm := info.Timestamp
		timeString := tm.Format("2006:01:02 15:04:05 -07:00")
		if tm.IsZero() || tm.Unix() == 0 {
			ifd, err = ifd.ChildWithIfdPath(exifcommon.IfdExifStandardIfdIdentity)
			if err != nil {
				fmt.Fprintf(os.Stderr, "No EXIF: %s\n", file)
				return nil, err
			}
			timeTags, err := ifd.FindTagWithName("DateTimeOriginal")
			if err != nil {
				fmt.Fprintf(os.Stderr, "No 'DateTimeOriginal': %s\n", file)
				return nil, err
			}
			offsetTags, err := ifd.FindTagWithName("OffsetTimeOriginal")
			if err != nil {
				fmt.Fprintf(os.Stderr, "No 'OffsetTimeOriginal': %s\n", file)
				return nil, err
			}
			if len(timeTags) == 1 && len(offsetTags) == 1 {
				datetime, err := timeTags[0].Format()
				if err != nil {
					return nil, err
				}
				offset, err := offsetTags[0].Format()
				if err != nil {
					return nil, err
				}
				timeString = fmt.Sprintf("%s %s", datetime, offset)
				tm, err = time.Parse("2006:01:02 15:04:05 -07:00", timeString)
				if err != nil {
					return nil, err
				}
			}
			fmt.Fprintf(os.Stderr, "%s: Timestamp from IFD/Exif\n", filepath.Base(file))
		} else {
			fmt.Fprintf(os.Stderr, "%s: Timestamp from IFD0/GPSInfo0\n", filepath.Base(file))
		}
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		digest := sha1.New()
		if _, err := io.Copy(digest, f); err != nil {
			return nil, err
		}
		f.Close()
		sum := fmt.Sprintf("%x", digest.Sum(nil))
		wpt := &gpx.WayPoint{
			Name:        proto.String(filepath.Base(file)),
			NanoTime:    proto.Int64(int64(tm.UnixNano())),
			Latitude:    proto.Float64(lat),
			Longitude:   proto.Float64(lng),
			Elevation:   proto.Float64(float64(alt)),
			Description: proto.String(fmt.Sprintf("Local time: %s\nsha1sum: %s", timeString, sum)),
		}
		waypoints = append(waypoints, wpt)
	}
	return waypoints, nil
}
