package display

import (
	"bytes"
	"image"

	"github.com/tinyzimmer/gsvnc/pkg/log"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
	"github.com/tinyzimmer/gsvnc/pkg/util"
)

// Server -> Client
const (
	encodingCopyRect     = 1
	cmdFramebufferUpdate = 0
)

func (d *Display) pushFrame(ur *types.FrameBufferUpdateRequest) {

	li := d.GetLastImage()
	if li == nil {
		return
	}

	if ur.Incremental() {
		li = truncateImage(ur, li)
	}

	log.Debug("Pushing latest frame to client")
	d.pushImage(li)
}

func (d *Display) pushImage(img *image.RGBA) {

	b := img.Bounds()

	buf := new(bytes.Buffer)

	util.Write(buf, uint8(cmdFramebufferUpdate))
	util.Write(buf, uint8(0))  // padding byte
	util.Write(buf, uint16(1)) // 1 rectangle

	//log.Printf("sending %d x %d pixels", width, height)
	format := d.GetPixelFormat()
	if format.TrueColour == 0 {
		log.Error("only true-colour supported")
		return
	}

	enc := d.GetCurrentEncoding()

	// Send that rectangle:
	util.PackStruct(buf, &types.FrameBufferRectangle{
		X: uint16(b.Min.X), Y: uint16(b.Min.Y), Width: uint16(b.Max.X - b.Min.X), Height: uint16(b.Max.Y - b.Min.Y), EncType: enc.Code(), // TODO make sure supported
	})

	enc.HandleBuffer(buf, d.GetPixelFormat(), img)

	d.buf.Dispatch(buf.Bytes())
}

func truncateImage(ur *types.FrameBufferUpdateRequest, img *image.RGBA) *image.RGBA {
	truncated := image.NewRGBA(
		image.Rect(
			int(ur.X), int(ur.Y), int(ur.Width), int(ur.Height),
		),
	)

	for y := ur.Y; y < ur.Height; y++ {
		for x := ur.X; x < ur.Width; x++ {
			truncated.Set(int(x), int(y), img.At(int(x), int(y)))
		}
	}

	return truncated
}
