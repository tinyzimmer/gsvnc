package encodings

import (
	"encoding/binary"
	"image"
	"io"
	"strconv"

	"github.com/tinyzimmer/go-gst/gst/video"

	"github.com/tinyzimmer/go-gst/gst"
)

// RawEncoding implements an Encoding intercace using raw pixel data.
type RawEncoding struct{}

// Code returns the code
func (r *RawEncoding) Code() int32 { return 0 }

// LinkPipeline links the pipeline.
func (r *RawEncoding) LinkPipeline(width, height int, pipeline *gst.Pipeline) (start, finish *gst.Element, err error) {
	elements, err := gst.NewElementMany("queue", "videoscale", "videorate", "videoconvert", "capsfilter")
	if err != nil {
		return nil, nil, err
	}

	pipeline.AddMany(elements...)

	videoInfo := video.NewInfo().
		WithFormat(video.FormatUYVY, uint(width), uint(height))
		// WithFPS(gst.Fraction(10, 1))
	// caps := gst.NewCapsFromString(fmt.Sprintf("video/x-raw, width=(int)%d, height=(int)%d", width, height))

	queue, videoscale, videorate, videoconvert, capsfilter :=
		elements[0], elements[1], elements[2], elements[3], elements[4]

	if err := gst.ElementLinkMany(queue, videoscale, videorate, videoconvert); err != nil {
		return nil, nil, err
	}

	return queue, capsfilter, videoconvert.LinkFiltered(capsfilter, videoInfo.ToCaps())
}

// HandleBuffer handles an image sample.
func (r *RawEncoding) HandleBuffer(w io.Writer, f *PixelFormat, buf []byte) {
	im := &image.RGBA{}
	im.Pix = buf
	b := im.Bounds()
	width, height := b.Dx(), b.Dy()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			col := im.At(x, y)
			r16, g16, b16, _ := col.RGBA()
			r16 = inRange(r16, f.RedMax)
			g16 = inRange(g16, f.GreenMax)
			b16 = inRange(b16, f.BlueMax)
			var u32 uint32 = (r16 << f.RedShift) |
				(g16 << f.GreenShift) |
				(b16 << f.BlueShift)
			var v interface{}
			switch f.BPP {
			case 32:
				v = u32
			case 16:
				v = uint16(u32)
			case 8:
				v = uint8(u32)
			default:
				return
			}
			if f.BigEndian != 0 {
				binary.Write(w, binary.BigEndian, v)
			} else {
				binary.Write(w, binary.LittleEndian, v)
			}
		}
	}
}

func inRange(v uint32, max uint16) uint32 {
	switch max {
	case 0x1f: // 5 bits
		return v >> (16 - 5)
	case 0xff:
		return v >> 8
	}
	panic("todo; max value = " + strconv.Itoa(int(max)))
}
