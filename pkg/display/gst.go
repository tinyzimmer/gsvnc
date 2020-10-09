package display

import (
	"log"
	"runtime"
	"time"

	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
	"github.com/tinyzimmer/gsvnc/pkg/encodings"
)

// Start will start the display. It assumes gstreamer has already been initialized.
func (d *Display) Start() error {
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return err
	}

	// Get the screen capture element depending on the OS
	src, err := d.getScreenCaptureElement()
	if err != nil {
		return err
	}

	decodebin, err := gst.NewElement("decodebin")
	if err != nil {
		return err
	}
	pipeline.AddMany(src, decodebin)
	src.Link(decodebin)

	decodebin.Connect("pad-added", func(self *gst.Element, srcPad *gst.Pad) {

		// Get the pipeline handler for the encoding we are going to use, looping until it is set.
		// It's the client job to let us know at some point. We give up after 10 seconds.
		var enc encodings.Encoding
		t := time.NewTicker(time.Second * 10)

	OuterLoop:
		for {
			select {
			case <-t.C:
				{
					return
				}
			default:
				enc = d.GetCurrentEncoding()
				if enc != nil {
					break OuterLoop
				}
			}
		}

		w, h := d.GetDimensions()

		// Let the encoding handler build out the rest of the pipeline
		start, finish, err := enc.LinkPipeline(w, h, pipeline)
		if err != nil {
			log.Println(err)
			return
		}

		// Don't bother the encoder with state syncing. Query the pipeline and do
		// it now.
		elements, err := pipeline.GetElements()
		if err != nil {
			log.Println(err)
			return
		}
		for _, e := range elements {
			e.SyncStateWithParent()
		}

		sink, err := app.NewAppSink()
		if err != nil {
			log.Println(err)
			return
		}

		pipeline.Add(sink.Element)
		finish.Link(sink.Element)

		sink.SetCallbacks(&app.SinkCallbacks{
			NewSampleFunc: func(self *app.Sink) gst.FlowReturn {
				// Pull the sample from the sink
				sample := self.PullSample()
				if sample == nil {
					return gst.FlowOK
				}
				defer sample.Unref()

				// Retrieve the pixels from the sample
				imgBytes := sample.GetBuffer().
					Map(gst.MapRead).
					Bytes()

				d.queue <- imgBytes
				return gst.FlowOK
			},
		})

		if ret := srcPad.Link(start.GetStaticPad("sink")); ret != gst.PadLinkOK {
			log.Println("Could not link src pad to pipeline")
		}
	})

	d.pipeline = pipeline
	return pipeline.SetState(gst.StatePlaying)
}

func (d *Display) getScreenCaptureElement() (elem *gst.Element, err error) {
	switch runtime.GOOS {

	case "windows":
		// Other option is to use directX
		elem, err = gst.NewElement("gdiscreencapsrc")
		if err != nil {
			return
		}
		// Defaults are satisfactory for starting

	case "darwin":
		// I think this is the only option for mac
		elem, err = gst.NewElement("avfvideosrc")
		if err != nil {
			return
		}
		elem.SetProperty("capture-screen", true)

	default:
		// For now the default assumes an X display
		elem, err = gst.NewElement("ximagesrc")
		if err != nil {
			return
		}
		elem.SetProperty("show-pointer", false)
		// XDamage will increase CPU usage considerably in some cases
		elem.SetProperty("use-damage", false)

	}
	return
}
