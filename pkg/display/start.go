package display

import (
	"bytes"
	"image"
	"runtime"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"

	"github.com/tinyzimmer/go-gst/gst"

	"github.com/tinyzimmer/gsvnc/pkg/log"
)

func logPipelineErr(err error) {
	log.Error("[go-gst-error] ", err.Error())
}

// Start will start a screen capture loop
func (d *Display) Start() error {
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
				w, h := d.GetDimensions()
				if b.Max.X > w || b.Max.Y > h {
					img = resize.Resize(uint(w), uint(h), img, resize.Lanczos3)
				}

				// if the image was resized this will be done already, otherwise, convert
				// to RGBA
				if _, ok := img.(*image.RGBA); !ok {
					img = convertToRGBA(img.(*image.NRGBA))
				}

				log.Debug("Queueing frame for processing")
				// Queue the image for processing
				select {
				case <-d.stopCh:
					log.Debug("Received event on stop channel, stopping screen capture")
					cont = false
				case d.frameQueue <- img.(*image.RGBA):
				default:
					// pop the oldest item off the queue
					// and let the next sample try to get in
					log.Debug("Client is behind on frames, forcing oldest one off the queue")
					select {
					case <-d.frameQueue:
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

// // Start will start the display. It assumes gstreamer has already been initialized.
// func (d *Display) Start() error {

// 	log.Debug("Building gstreamer pipeline for display connection")

// 	pipeline, err := gst.NewPipeline("")
// 	if err != nil {
// 		return err
// 	}

// 	// Get the screen capture element depending on the OS
// 	src, err := d.getScreenCaptureElement()
// 	if err != nil {
// 		return err
// 	}

// 	// Let decodebin decide the best pipelin depending on the source stream
// 	decodebin, err := gst.NewElement("decodebin")
// 	if err != nil {
// 		return err
// 	}
// 	pipeline.AddMany(src, decodebin)
// 	src.Link(decodebin)

// 	// Buiild out the rest of the pipeline when decodebin is ready.
// 	decodebin.Connect("pad-added", func(self *gst.Element, srcPad *gst.Pad) {
// 		log.Debug("Decodebin pad added, linking pipeline")
// 		elements, err := gst.NewElementMany("queue", "videorate", "videoscale", "videoconvert", "jpegenc", "appsink")
// 		if err != nil {
// 			logPipelineErr(err)
// 			return
// 		}

// 		queue, videorate, videoscale, videoconvert, jpegenc, appsink :=
// 			elements[0], elements[1], elements[2], elements[3], elements[4], elements[5]

// 		// Build out caps
// 		w, h := d.GetDimensions()
// 		rateCaps := gst.NewCapsFromString("video/x-raw, framerate=10/1")
// 		scaleCaps := gst.NewCapsFromString(fmt.Sprintf("video/x-raw, width=%d, height=%d", w, h))
// 		videoInfo := video.NewInfo().
// 			WithFormat(video.FormatRGBx, uint(w), uint(h)).
// 			WithFPS(gst.Fraction(10, 1))

// 		// Configure and link elements
// 		if err := runAllUntilError([]func() error{
// 			func() error { return videoscale.SetProperty("sharpen", 1) },
// 			func() error { return videoscale.SetProperty("method", 0) }, // Use nearest neighbor - significantly cheaper CPU
// 			func() error { return pipeline.AddMany(elements...) },
// 			func() error { return queue.Link(videorate) },
// 			func() error { return videorate.LinkFiltered(videoscale, rateCaps) },
// 			func() error { return videoscale.LinkFiltered(videoconvert, scaleCaps) },
// 			func() error { return videoconvert.LinkFiltered(jpegenc, videoInfo.ToCaps()) },
// 			func() error { return jpegenc.Link(appsink) },
// 		}); err != nil {
// 			logPipelineErr(err)
// 			return
// 		}

// 		log.Debug("Syncing new element states with parent pipeline")
// 		for _, e := range elements {
// 			if ok := e.SyncStateWithParent(); !ok {
// 				logPipelineErr(fmt.Errorf("Could not sink element state with parent: %s", e.Name()))
// 				return
// 			}
// 		}

// 		// Connect to new samples on the sink
// 		sink := app.SinkFromElement(appsink)
// 		sink.SetCallbacks(&app.SinkCallbacks{
// 			NewSampleFunc: func(self *app.Sink) gst.FlowReturn {
// 				// Pull the sample from the sink
// 				sample := self.PullSample()
// 				if sample == nil {
// 					return gst.FlowOK
// 				}
// 				defer sample.Unref()

// 				log.Debug("Received new frame on the pipeline, decoding")

// 				// Retrieve the pixels from the sample
// 				buf := sample.GetBuffer().Reader()

// 				img, err := jpeg.Decode(buf)
// 				if err != nil {
// 					logPipelineErr(err)
// 					return gst.FlowError
// 				}

// 				log.Debug("Queueing frame for processing")
// 				// Queue the image for processing
// 				var ok bool
// 				select {
// 				case d.frameQueue <- img.(*image.RGBA):
// 					ok = true
// 				default:
// 					ok = false
// 					// pop the oldest item off the queue
// 					// and let the next sample try to get in
// 					select {
// 					case <-d.frameQueue:
// 					}
// 				}

// 				if !ok {
// 					log.Warning("Client is behind on frames, could not push to channel")
// 				} else {
// 					log.Debug("Successfully queued frame for processing")
// 				}

// 				return gst.FlowOK
// 			},
// 		})

// 		if ret := srcPad.Link(queue.GetStaticPad("sink")); ret != gst.PadLinkOK {
// 			log.Error("Could not link src pad to pipeline")
// 		}
// 	})

// 	if config.Debug {
// 		bus := pipeline.GetPipelineBus()
// 		go func() {
// 			for {
// 				msg := bus.TimedPop(time.Duration(-1))
// 				if msg == nil {
// 					return
// 				}
// 				log.Debug(msg)
// 				msg.Unref()
// 			}
// 		}()
// 	}

// 	d.pipeline = pipeline
// 	return pipeline.SetState(gst.StatePlaying)
// }

func (d *Display) getScreenCaptureElement() (elem *gst.Element, err error) {
	switch runtime.GOOS {

	case "windows":
		log.Debug("Detected Windows, using gdiscreencapsrc")
		// Other option is to use directX
		elem, err = gst.NewElement("gdiscreencapsrc")
		if err != nil {
			return
		}
		err = elem.SetProperty("cursor", true)

	case "darwin":
		log.Debug("Detected macOS, using avfvideosrc")
		// I think this is the only option for mac
		elem, err = gst.NewElement("avfvideosrc")
		if err != nil {
			return
		}
		err = elem.SetProperty("capture-screen", true)
		if err != nil {
			return
		}
		err = elem.SetProperty("capture-screen-cursor", true)

	default:
		log.Debug("Detected Linux, using ximagesrc")
		// For now the default assumes an X display
		elem, err = gst.NewElement("ximagesrc")
		if err != nil {
			return
		}
		err = elem.SetProperty("show-pointer", true)
		if err != nil {
			return
		}
		// XDamage will increase CPU usage considerably in some cases
		err = elem.SetProperty("use-damage", false)

	}
	return
}

func runAllUntilError(fs []func() error) error {
	for _, f := range fs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}
