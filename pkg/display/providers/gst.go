package providers

import (
	"fmt"
	"image"
	"image/jpeg"
	"runtime"
	"time"

	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
	"github.com/tinyzimmer/go-gst/gst/video"

	"github.com/tinyzimmer/gsvnc/pkg/config"
	"github.com/tinyzimmer/gsvnc/pkg/internal/log"
)

// Gstreamer implements a display provider using gstreamer to capture
// video from the display.
type Gstreamer struct {
	pipeline   *gst.Pipeline
	frameQueue chan *image.RGBA // A channel that will essentially only ever have the latest frame available.
}

// Close stops the gstreamer pipeline.
func (g *Gstreamer) Close() error { return g.pipeline.Destroy() }

// PullFrame returns a frame from the queue.
func (g *Gstreamer) PullFrame() *image.RGBA { return <-g.frameQueue }

// Start will start the gstreamer pipelines and send imags to the frame queue.
func (g *Gstreamer) Start(width, height int) error {
	log.Debug("Building gstreamer pipeline for display connection")
	g.frameQueue = make(chan *image.RGBA, 2)

	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return err
	}

	// Get the screen capture element depending on the OS
	src, err := getScreenCaptureElement()
	if err != nil {
		return err
	}

	// Let decodebin decide the best pipelin depending on the source stream
	decodebin, err := gst.NewElement("decodebin")
	if err != nil {
		return err
	}
	pipeline.AddMany(src, decodebin)
	src.Link(decodebin)

	// Buiild out the rest of the pipeline when decodebin is ready.
	decodebin.Connect("pad-added", func(self *gst.Element, srcPad *gst.Pad) {
		log.Debug("Decodebin pad added, linking pipeline")
		elements, err := gst.NewElementMany("queue", "videorate", "videoscale", "videoconvert", "jpegenc", "appsink")
		if err != nil {
			logPipelineErr(err)
			return
		}

		queue, videorate, videoscale, videoconvert, jpegenc, appsink :=
			elements[0], elements[1], elements[2], elements[3], elements[4], elements[5]

		// Build out caps
		rateCaps := gst.NewCapsFromString("video/x-raw, framerate=5/1")
		scaleCaps := gst.NewCapsFromString(fmt.Sprintf("video/x-raw, width=%d, height=%d", width, height))
		videoInfo := video.NewInfo().
			WithFormat(video.FormatRGBx, uint(width), uint(height)).
			WithFPS(gst.Fraction(5, 1))

		// Configure and link elements
		if err := runAllUntilError([]func() error{
			func() error { return videoscale.SetProperty("sharpen", 1) },
			func() error { return videoscale.SetProperty("method", 0) }, // Use nearest neighbor - significantly cheaper CPU
			func() error { return pipeline.AddMany(elements...) },
			func() error { return queue.Link(videorate) },
			func() error { return videorate.LinkFiltered(videoscale, rateCaps) },
			func() error { return videoscale.LinkFiltered(videoconvert, scaleCaps) },
			func() error { return videoconvert.LinkFiltered(jpegenc, videoInfo.ToCaps()) },
			func() error { return jpegenc.Link(appsink) },
		}); err != nil {
			logPipelineErr(err)
			return
		}

		log.Debug("Syncing new element states with parent pipeline")
		for _, e := range elements {
			if ok := e.SyncStateWithParent(); !ok {
				logPipelineErr(fmt.Errorf("Could not sink element state with parent: %s", e.Name()))
				return
			}
		}

		// Connect to new samples on the sink
		sink := app.SinkFromElement(appsink)
		sink.SetCallbacks(&app.SinkCallbacks{
			NewSampleFunc: func(self *app.Sink) gst.FlowReturn {
				// Pull the sample from the sink
				sample := self.PullSample()
				if sample == nil {
					return gst.FlowOK
				}
				defer sample.Unref()

				log.Debug("Received new frame on the pipeline, decoding")

				// Retrieve the pixels from the sample
				buf := sample.GetBuffer().Reader()

				img, err := jpeg.Decode(buf)
				if err != nil {
					logPipelineErr(err)
					return gst.FlowError
				}

				log.Debug("Queueing frame for processing")
				// Queue the image for processing
				var ok bool
				select {
				case g.frameQueue <- img.(*image.RGBA):
					ok = true
				default:
					ok = false
					// pop the oldest item off the queue
					// and let the next sample try to get in
					select {
					case <-g.frameQueue:
					}
				}

				if !ok {
					log.Debug("Client is behind on frames, could not push to channel")
				} else {
					log.Debug("Successfully queued frame for processing")
				}

				return gst.FlowOK
			},
		})

		if ret := srcPad.Link(queue.GetStaticPad("sink")); ret != gst.PadLinkOK {
			log.Error("Could not link src pad to pipeline")
		}
	})

	if config.Debug {
		bus := pipeline.GetPipelineBus()
		go func() {
			for {
				msg := bus.TimedPop(time.Duration(-1))
				if msg == nil {
					return
				}
				log.Debug(msg)
				msg.Unref()
			}
		}()
	}

	g.pipeline = pipeline
	return pipeline.SetState(gst.StatePlaying)
}

func getScreenCaptureElement() (elem *gst.Element, err error) {
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

func logPipelineErr(err error) {
	log.Error("[go-gst-error] ", err.Error())
}
