package encodings

import (
	"image"
	"io"
	"reflect"

	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

// Encoding is an interface to be implemented by different encoding handlers.
type Encoding interface {
	// Code should return the int32 code of the encoding type.
	Code() int32
	// HandleBuffer should craft a rectangle from the given image and
	// queue it onto the given writer.
	HandleBuffer(w io.Writer, format *types.PixelFormat, img *image.RGBA)
}

// EnabledEncodings lists the encodings currently enabled by the server. It can be mutated
// by command line options.
var EnabledEncodings = []Encoding{
	&RawEncoding{},
	&TightEncoding{},
	&TightPNGEncoding{},
}

// GetDefaults returns a slice of the default encoding handlers.
func GetDefaults() []Encoding {
	out := make([]Encoding, len(EnabledEncodings))
	for i, t := range EnabledEncodings {
		out[i] = t
	}
	return out
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
