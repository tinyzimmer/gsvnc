package encodings

import (
	"io"

	"github.com/tinyzimmer/go-gst/gst/video"

	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/gsvnc/pkg/util"
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
func (r *RawEncoding) HandleBuffer(w io.Writer, buf []byte) {
	util.Write(w, buf)
}
