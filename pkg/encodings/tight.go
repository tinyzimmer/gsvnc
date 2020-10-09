package encodings

import (
	"encoding/binary"
	"io"
	"strconv"

	"github.com/tinyzimmer/go-gst/gst/video"

	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
	"github.com/tinyzimmer/gsvnc/pkg/util"
)

// TightEncoding implements an Encoding intercace using Tight encoding.
type TightEncoding struct{}

// Code returns the code
func (t *TightEncoding) Code() int32 { return 7 }

// LinkPipeline links the pipeline.
func (t *TightEncoding) LinkPipeline(width, height int, pipeline *gst.Pipeline) (start, finish *gst.Element, err error) {
	elements, err := gst.NewElementMany("queue", "videoscale", "videorate", "videoconvert", "jpegenc", "jifmux")
	if err != nil {
		return nil, nil, err
	}

	pipeline.AddMany(elements...)

	videoInfo := video.NewInfo().
		WithFormat(video.FormatUYVY, uint(width), uint(height))
		// WithFPS(gst.Fraction(10, 1))
	// caps := gst.NewCapsFromString(fmt.Sprintf("video/x-raw, width=(int)%d, height=(int)%d", width, height))

	queue, videoscale, videorate, videoconvert, jpegenc, jifmux :=
		elements[0], elements[1], elements[2], elements[3], elements[4], elements[5]

	if err := gst.ElementLinkMany(queue, videoscale, videorate, videoconvert); err != nil {
		return nil, nil, err
	}
	if err := videoconvert.LinkFiltered(jpegenc, videoInfo.ToCaps()); err != nil {
		return nil, nil, err
	}

	return queue, jifmux, jpegenc.Link(jifmux)
}

// HandleBuffer handles an image sample.
func (t *TightEncoding) HandleBuffer(w io.Writer, f *types.PixelFormat, buf []byte) {
	i, _ := strconv.ParseInt("10010000", 2, 64) // JPEG encoding
	util.Write(w, uint8(i))

	// Buffer length
	util.Write(w, computeTightLength(len(buf)))

	// Buffer contents
	util.Write(w, buf)
}

func computeTightLength(num int) (b []byte) {
	var out []byte
	if num < 127 {
		out = make([]byte, 1)
	} else if num > 127 && num < 16383 {
		out = make([]byte, 2)
	} else if num > 16383 {
		out = make([]byte, 3)
	}
	binary.PutUvarint(out, uint64(num))
	return out
}
