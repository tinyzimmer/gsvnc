package encodings

import (
	"io"
	"reflect"

	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

// Encoding is an interface to be implemented by different encoding handlers.
type Encoding interface {
	// Code should return the int32 code of the encoding type.
	Code() int32
	// LinkPipeline should build out the elements the encoding needs into the
	// given pipeline. The buffers produced by the pipeline will be passed to
	// the HandleBuffer handler.
	//
	// The returned element must be the element that can be linked to the source
	// and sink on the pipeline. This (as well as syncting state with parent) is handled
	// externally to this function.
	LinkPipeline(width, height int, pipeline *gst.Pipeline) (start, end *gst.Element, err error)
	// HandleBuffer should craft a rectangle from the given byte sequence and
	// queue it onto the given writer.
	HandleBuffer(w io.Writer, format *types.PixelFormat, buf []byte)
}

// EnabledEncodings lists the encodings currently enabled by the server. It can be mutated
// by command line options.
var EnabledEncodings = []Encoding{
	&RawEncoding{},
	&TightEncoding{},
}

// GetEncoding will iterate the requested encodings and return the best match
// that can be served. If none of the requested encodings are supported (should
// never happen as at least RAW is required by RFC) this function returns nil.
func GetEncoding(encs []int32) Encoding {
	for _, e := range encs {
		for _, supported := range EnabledEncodings {
			if e == supported.Code() {
				return supported
			}
		}
	}
	return nil
}

// DisableEncoding removes the given encoding from the list of EnabledEncodings.
func DisableEncoding(enc Encoding) {
	EnabledEncodings = remove(EnabledEncodings, enc)
}

func remove(tt []Encoding, t Encoding) []Encoding {
	newTypes := make([]Encoding, 0)
	for _, enabled := range tt {
		if reflect.TypeOf(enabled).Elem().Name() != reflect.TypeOf(t).Elem().Name() {
			newTypes = append(newTypes, enabled)
		}
	}
	return newTypes
}
