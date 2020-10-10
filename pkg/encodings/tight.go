package encodings

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"log"
	"strconv"

	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
	"github.com/tinyzimmer/gsvnc/pkg/util"
)

// TightEncoding implements an Encoding intercace using Tight encoding.
type TightEncoding struct{}

// Code returns the code
func (t *TightEncoding) Code() int32 { return 7 }

// HandleBuffer handles an image sample.
func (t *TightEncoding) HandleBuffer(w io.Writer, f *types.PixelFormat, img *image.RGBA) {
	compressed := new(bytes.Buffer)

	err := jpeg.Encode(compressed, img, nil)
	if err != nil {
		log.Println("[tight-jpeg] Could not encode image frame to jpeg")
		return
	}

	buf := compressed.Bytes()

	i, _ := strconv.ParseInt("10010000", 2, 64) // JPEG encoding
	util.Write(w, uint8(i))

	// Buffer length
	util.Write(w, computeTightLength(len(buf)))

	// Buffer contents
	util.Write(w, buf)
}

func computeTightLength(compressedLen int) (b []byte) {
	out := []byte{byte(compressedLen & 0x7F)}
	if compressedLen > 0x7F {
		out[0] |= 0x80
		out = append(out, byte(compressedLen>>7&0x7F))
		if compressedLen > 0x3FFF {
			out[1] |= 0x80
			out = append(out, byte(compressedLen>>14&0xFF))
		}
	}
	return out
}
