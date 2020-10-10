package encodings

import (
	"bytes"
	"encoding/binary"
	"image"
	"strconv"

	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

func applyPixelFormat(img *image.RGBA, format *types.PixelFormat) []byte {
	formattedImage := new(bytes.Buffer)
	b := img.Bounds()
	width, height := b.Dx(), b.Dy()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			col := img.At(x, y)
			r16, g16, b16, _ := col.RGBA()
			r16 = inRange(r16, format.RedMax)
			g16 = inRange(g16, format.GreenMax)
			b16 = inRange(b16, format.BlueMax)
			var u32 uint32 = (r16 << format.RedShift) |
				(g16 << format.GreenShift) |
				(b16 << format.BlueShift)
			var v interface{}
			switch format.BPP {
			case 32:
				v = u32
			case 16:
				v = uint16(u32)
			case 8:
				v = uint8(u32)
			}
			if format.BigEndian != 0 {
				binary.Write(formattedImage, binary.BigEndian, v)
			} else {
				binary.Write(formattedImage, binary.LittleEndian, v)
			}
		}
	}
	return formattedImage.Bytes()
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
