package encodings

import (
	"io"
	"reflect"

	"github.com/tinyzimmer/go-gst/gst"
)

// PixelFormat represents the current pixel format for the display.
// Fields are declared in the order they appear in a SetPixelFormat
// request.
type PixelFormat struct {
	BPP, Depth                      uint8
	BigEndian, TrueColour           uint8 // flags; 0 or non-zero
	RedMax, GreenMax, BlueMax       uint16
	RedShift, GreenShift, BlueShift uint8
}

// IsScreensThousands returns if the format requested by the OS X "Screens" app's "Thousands" mode.
func (f *PixelFormat) IsScreensThousands() bool {
	// Note: Screens asks for Depth 16; RealVNC asks for Depth 15 (which is more accurate)
	// Accept either. Same format.
	return f.BPP == 16 && (f.Depth == 16 || f.Depth == 15) && f.TrueColour != 0 &&
		f.RedMax == 0x1f && f.GreenMax == 0x1f && f.BlueMax == 0x1f &&
		f.RedShift == 10 && f.GreenShift == 5 && f.BlueShift == 0
}

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
	HandleBuffer(w io.Writer, format *PixelFormat, buf []byte)
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
