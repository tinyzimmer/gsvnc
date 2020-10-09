package display

import (
	"bytes"
	"log"
	"time"

	"github.com/tinyzimmer/gsvnc/pkg/util"
)

// FrameBufferUpdateRequest represents a request to update the frame buffer.
type FrameBufferUpdateRequest struct {
	IncrementalFlag     uint8
	X, Y, Width, Height uint16
}

func (r *FrameBufferUpdateRequest) incremental() bool { return r.IncrementalFlag != 0 }

func (d *Display) pushFramesLoop() {
	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case ur, ok := <-d.reqQueue:
			if !ok {
				// Client disconnected.
				return
			}
			d.pushFrame(ur)
		case <-ticker.C:
			last := d.GetLastImage()
			d.pushImage(last)
		}
	}
}

// Server -> Client
const (
	encodingRaw          = 0
	encodingCopyRect     = 1
	cmdFramebufferUpdate = 0
)

type frameBufferRectangle struct {
	X, Y          uint16
	Width, Height uint16
	EncType       int32
}

func (d *Display) pushFrame(ur *FrameBufferUpdateRequest) {

	li := d.GetLastImage()
	if li == nil {
		return
	}

	if ur.incremental() {
		width, height := d.GetDimensions()
		buf := new(bytes.Buffer)

		// log.Printf("Client wants incremental update, sending none. %#v", ur)

		util.Write(buf, uint8(cmdFramebufferUpdate))
		// padding byte
		util.Write(buf, uint8(0))
		// num rectangles
		util.Write(buf, uint16(1))

		util.PackStruct(buf, &frameBufferRectangle{
			X: 0, Y: 0, Width: uint16(width), Height: uint16(height), EncType: encodingCopyRect, // TODO make sure supported
		})

		util.Write(buf, uint16(0)) // src-x
		util.Write(buf, uint16(0)) // src-y

		d.buf.Dispatch(buf.Bytes())
		return
	}

	d.pushImage(li)
}

func (d *Display) pushImage(imgData []byte) {

	width, height := d.GetDimensions()

	buf := new(bytes.Buffer)

	util.Write(buf, uint8(cmdFramebufferUpdate))
	util.Write(buf, uint8(0))  // padding byte
	util.Write(buf, uint16(1)) // 1 rectangle

	//log.Printf("sending %d x %d pixels", width, height)
	format := d.GetPixelFormat()
	if format.TrueColour == 0 {
		log.Println("only true-colour supported")
		return
	}

	enc := d.GetCurrentEncoding()

	// Send that rectangle:
	util.PackStruct(buf, &frameBufferRectangle{
		X: 0, Y: 0, Width: uint16(width), Height: uint16(height), EncType: enc.Code(), // TODO make sure supported
	})

	enc.HandleBuffer(buf, imgData)

	d.buf.Dispatch(buf.Bytes())
}
