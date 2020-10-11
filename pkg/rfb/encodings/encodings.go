package encodings

import (
	"image"
	"io"

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

// DefaultEncodings lists the encodings enabled by default on the server.
var DefaultEncodings = []Encoding{
	&RawEncoding{},
	&TightEncoding{},
	&TightPNGEncoding{},
}

// GetDefaults returns a slice of the default encoding handlers.
func GetDefaults() []Encoding {
	out := make([]Encoding, len(DefaultEncodings))
	for i, t := range DefaultEncodings {
		out[i] = t
	}
	return out
}
