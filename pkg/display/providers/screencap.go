package providers

import (
	"bytes"
	"image"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"

	"github.com/tinyzimmer/gsvnc/pkg/internal/log"
)

// ScreenCapture implements a display provider that periodically captures the screen
// using native APIs.
type ScreenCapture struct {
	frameQueue chan *image.RGBA // A channel that will essentially only ever have the latest frame available.
	stopCh     chan struct{}
}

// Close stops the gstreamer pipeline.
func (s *ScreenCapture) Close() error {
	s.stopCh <- struct{}{}
	return nil
}

// PullFrame returns a frame from the queue.
func (s *ScreenCapture) PullFrame() *image.RGBA { return <-s.frameQueue }

// Start starts the screen capture loop.
func (s *ScreenCapture) Start(width, height int) error {
	s.frameQueue = make(chan *image.RGBA, 2)
	s.stopCh = make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond * 200) // 5 frames a second
		for range ticker.C {
			cont := true

			func() {
				bitMap := robotgo.CaptureScreen()
				defer robotgo.FreeBitmap(bitMap)

				bs := robotgo.ToBitmapBytes(bitMap)

				var img image.Image

				img, err := bmp.Decode(bytes.NewReader(bs))
				if err != nil {
					log.Error("Unable to decode bitmap: ", err.Error())
					return
				}

				b := img.Bounds()
				if b.Max.X > width || b.Max.Y > height {
					img = resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
				}

				// if the image was resized this will be done already, otherwise, convert
				// to RGBA
				if _, ok := img.(*image.RGBA); !ok {
					img = convertToRGBA(img.(*image.NRGBA))
				}

				log.Debug("Queueing frame for processing")
				// Queue the image for processing
				select {
				case <-s.stopCh:
					log.Debug("Received event on stop channel, stopping screen capture")
					cont = false
				case s.frameQueue <- img.(*image.RGBA):
				default:
					// pop the oldest item off the queue
					// and let the next sample try to get in
					log.Debug("Client is behind on frames, forcing oldest one off the queue")
					select {
					case <-s.frameQueue:
					}
				}

			}()

			if !cont {
				return
			}
		}
	}()
	return nil
}

func convertToRGBA(in *image.NRGBA) *image.RGBA {
	size := in.Bounds().Size()
	rect := image.Rect(0, 0, size.X, size.Y)
	wImg := image.NewRGBA(rect)
	// loop though all the x
	for x := 0; x < size.X; x++ {
		// and now loop thorough all of this x's y
		for y := 0; y < size.Y; y++ {
			wImg.Set(x, y, in.At(x, y))
		}
	}
	return wImg
}
