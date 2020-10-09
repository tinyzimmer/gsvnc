package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/go-vgo/robotgo"
	"github.com/spf13/cobra"

	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/gsvnc/pkg/encodings"
	"github.com/tinyzimmer/gsvnc/pkg/rfb"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/auth"
)

var bindHost string
var bindPort int32
var initialResolution string
var listFeatures bool

var rootCmd = &cobra.Command{
	Use:   "gsvnc",
	Short: "Gsvnc is an extensible, cross-platform VNC server written in go.",
	Long: `Gsvnc is intended to be a fast and flexible VNC server, devoid of the complexities of the many out there written in C.
It uses gstreamer on the backend to provide framebuffer (and eventually audio via QEMU extensions) streams to connected clients.

The supported security/encoding types are limited at the moment, but the intention is to implement at least all of the core ones.
Then, either provide a pluggable interface for implementing optional features, or at least keep the code base simple enough to make
implementing them easy.

By default only core security types and encodings are enabled, however you can disable/enable different features by using the 
+/- syntax at the end of the command line. For example:

	gsvnc -- +TightSecurity -None -RawEncoding +TightEncoding

A list of all available features and their default status can be obtained with --list-features.
   (You can also use this command to see the effect of the positional flags)
`,
	RunE: run,
}

func init() {
	gst.Init(nil)

	rootCmd.PersistentFlags().StringVarP(&bindHost, "host", "H", "127.0.0.1", "The host address to bind the server to.")
	rootCmd.PersistentFlags().Int32VarP(&bindPort, "port", "p", 5900, "The port to bind the server to.")
	rootCmd.PersistentFlags().StringVarP(&initialResolution, "resolution", "r", "", "The initial resolution to set for display connections. Defaults to auto-detect.")
	rootCmd.PersistentFlags().BoolVarP(&listFeatures, "list-features", "l", false, "List the available features and exit.")
}

func run(cmd *cobra.Command, args []string) error {

	if err := configureFeatures(args); err != nil {
		return err
	}

	if listFeatures {
		doListFeatures()
		os.Exit(0)
	}

	log.Println("Starting gsvnc")

	bindAddr := fmt.Sprintf("%s:%d", bindHost, bindPort)

	// Create a listener
	l, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return err
	}

	var w, h int

	if initialResolution == "" {
		w, h = robotgo.GetScreenSize()
		log.Printf("Detected initial screen resolution of %dx%d", w, h)
	} else {
		spl := strings.Split(strings.ToLower(initialResolution), "x")
		if len(spl) != 2 {
			return fmt.Errorf("Could not parse provided resolution: %s", initialResolution)
		}
		w, err = strconv.Atoi(spl[0])
		if err != nil {
			return fmt.Errorf("Could not parse '%s' as an integer", spl[0])
		}
		h, err = strconv.Atoi(spl[1])
		if err != nil {
			return fmt.Errorf("Could not parse '%s' as an integer", spl[1])
		}
		log.Printf("Using initial screen resolution of %dx%d", w, h)
	}

	var enabledAuths, enabledEncs []string
	for _, sec := range auth.EnabledAuthTypes {
		enabledAuths = append(enabledAuths, reflect.TypeOf(sec).Elem().Name())
	}
	for _, enc := range encodings.EnabledEncodings {
		enabledEncs = append(enabledEncs, reflect.TypeOf(enc).Elem().Name())
	}
	log.Println("Enabled security types:", enabledAuths)
	log.Println("Enabled encodings:", enabledEncs)

	// Create a new rfb server
	server := rfb.NewServer(w, h)

	// Start the server
	log.Println("Listening for rfb connections on", bindAddr)

	return server.Serve(l)
}

func configureFeatures(args []string) error {
	if len(args) == 0 {
		return nil
	}
ArgLoop:
	for _, arg := range args {

		if strings.HasPrefix(arg, "+") {

			fmt.Println("There are currently no optional featuress to enable, ignoring", arg)

		} else if strings.HasPrefix(arg, "-") {

			featName := strings.TrimPrefix(arg, "-")

			for _, sec := range auth.EnabledAuthTypes {
				if reflect.TypeOf(sec).Elem().Name() == featName {
					auth.DisableAuth(sec)
					continue ArgLoop
				}
			}
			for _, enc := range encodings.EnabledEncodings {
				if reflect.TypeOf(enc).Elem().Name() == featName {
					encodings.DisableEncoding(enc)
					continue ArgLoop
				}
			}

			// If we got here it means we didn't have anything matching the request.
			return fmt.Errorf("Could not find any features with the name, %s", featName)

		} else {
			return fmt.Errorf("Bogus argument: %s", arg)
		}
	}
	return nil
}

func doListFeatures() {
	w := new(tabwriter.Writer)
	buf := new(bytes.Buffer)

	w.Init(
		buf,
		20,  // minwidth
		30,  // tabwith
		0,   // padding
		' ', // padchar
		0,   // flags
	)

	fmt.Fprintln(w, "Security Types")
	fmt.Fprintln(w, "--------------")
	for _, sec := range auth.EnabledAuthTypes {
		fmt.Fprintf(w, "%s\t(enabled)\n", reflect.TypeOf(sec).Elem().Name())
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Encodings")
	fmt.Fprintln(w, "---------")
	for _, enc := range encodings.EnabledEncodings {
		fmt.Fprintf(w, "%s\t(enabled)\n", reflect.TypeOf(enc).Elem().Name())
	}
	w.Flush()
	fmt.Println(buf.String())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
