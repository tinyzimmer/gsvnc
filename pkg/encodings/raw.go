package encodings

import (
	"image"
	"io"

	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

// RawEncoding implements an Encoding intercace using raw pixel data.
type RawEncoding struct{}

// Code returns the code
func (r *RawEncoding) Code() int32 { return 0 }

// HandleBuffer handles an image sample.
func (r *RawEncoding) HandleBuffer(w io.Writer, f *types.PixelFormat, img *image.NRGBA) {
	w.Write(applyPixelFormat(img, f))
}
