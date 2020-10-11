package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/go-vgo/robotgo"
	"github.com/spf13/cobra"

	"github.com/tinyzimmer/go-gst/gst"

	"github.com/tinyzimmer/gsvnc/pkg/config"
	"github.com/tinyzimmer/gsvnc/pkg/display/providers"
	"github.com/tinyzimmer/gsvnc/pkg/internal/log"
	"github.com/tinyzimmer/gsvnc/pkg/internal/util"
	"github.com/tinyzimmer/gsvnc/pkg/rfb"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/auth"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/encodings"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/events"
)

var bindHost string
var bindPort int32
var initialResolution string
var listFeatures bool
var displayProvider string
var websockify bool
var websockifyHost string
var websockifyPort int32
var noTCP bool
var serverPasswordFile string

// RootCmd is the exported root cmd for the gsvnc server.
var RootCmd = &cobra.Command{
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

	RootCmd.PersistentFlags().StringVarP(&bindHost, "host", "H", "127.0.0.1", "The host address to bind the server to.")
	RootCmd.PersistentFlags().Int32VarP(&bindPort, "port", "p", 5900, "The port to bind the server to.")
	RootCmd.PersistentFlags().StringVarP(&initialResolution, "resolution", "r", "", "The initial resolution to set for display connections. Defaults to auto-detect.")
	RootCmd.PersistentFlags().StringVarP(&serverPasswordFile, "password-file", "", "", "A file to read in a server password from. One will be generated if this is omitted.")
	RootCmd.PersistentFlags().BoolVarP(&listFeatures, "list-features", "l", false, "List the available features and exit.")
	RootCmd.PersistentFlags().StringVarP(&displayProvider, "display", "D", providers.ProviderGstreamer, "The display provider to use for RFB connections.")
	RootCmd.PersistentFlags().BoolVarP(&websockify, "websockify", "w", false, "Start a websockify listener")
	RootCmd.PersistentFlags().StringVarP(&websockifyHost, "websockify-host", "W", "127.0.0.1", "The host address to bind the websockify server to.")
	RootCmd.PersistentFlags().Int32VarP(&websockifyPort, "websockify-port", "P", 8080, "The port to bind the websockify server to.")
	RootCmd.PersistentFlags().BoolVarP(&noTCP, "no-tcp", "T", false, "Disable the TCP listener. Only makes sense with --websockify.")
	RootCmd.PersistentFlags().BoolVarP(&config.Debug, "debug", "d", false, "Enable debug logging.")
}

func run(cmd *cobra.Command, args []string) error {

	var err error

	authTypes, encTypes, eventTypes := configureFeatures(args)

	if listFeatures {
		doListFeatures(authTypes, encTypes, eventTypes)
		os.Exit(0)
	}

	log.Info("Starting gsvnc")

	// Make sure the configured display provider is valid.
	if p := providers.GetDisplayProvider(providers.Provider(displayProvider)); p == nil {
		return fmt.Errorf("Display provider is invalid: %s", displayProvider)
	}
	log.Info("Using display provider: ", displayProvider)

	// Configure initial display resolution
	var w, h int
	if initialResolution == "" {
		w, h = robotgo.GetScreenSize()
		log.Infof("Detected initial screen resolution of %dx%d", w, h)
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
		log.Infof("Using initial screen resolution of %dx%d", w, h)
	}

	var enabledAuths, enabledEncs, enabledEvents []string
	for _, sec := range authTypes {
		enabledAuths = append(enabledAuths, reflect.TypeOf(sec).Elem().Name())
	}
	for _, enc := range encTypes {
		enabledEncs = append(enabledEncs, reflect.TypeOf(enc).Elem().Name())
	}
	for _, ev := range eventTypes {
		enabledEvents = append(enabledEvents, reflect.TypeOf(ev).Elem().Name())
	}

	log.Info("Enabled security types: ", enabledAuths)
	log.Info("Enabled encodings: ", enabledEncs)
	log.Info("Enabled event handlers: ", enabledEvents)

	opts := &rfb.ServerOpts{
		Width: w, Height: h,
		DisplayProvider:  providers.Provider(displayProvider),
		EnabledAuthTypes: authTypes,
		EnabledEncodings: encTypes,
		EnabledEvents:    eventTypes,
	}

	if authIsEnabled(authTypes, "VNCAuth") {
		if serverPasswordFile != "" {
			passw, err := ioutil.ReadFile(serverPasswordFile)
			if err != nil {
				return err
			}
			opts.ServerPassword = string(passw)
		} else {
			log.Info("VNCAuth is enabled and no password provided, generating a server password")
			opts.ServerPassword = util.RandomString(8)
			log.Info("Clients using VNCAuth can connect with the following password: ", opts.ServerPassword)
		}
	}

	// Create a new rfb server
	server := rfb.NewServer(opts)

	if noTCP && !websockify {
		return errors.New("No listeners configured")
	}

	if noTCP && websockify {
		// We are only doing websockify
		return serveWebsockify(server)
	}

	if websockify {
		go serveWebsockify(server)
	}

	// Create a listener
	bindAddr := fmt.Sprintf("%s:%d", bindHost, bindPort)
	l, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return err
	}
	log.Info("Listening for rfb connections on ", bindAddr)
	return server.Serve(l)
}

func serveWebsockify(srvr *rfb.Server) error {
	wsAddr := fmt.Sprintf("%s:%d", websockifyHost, websockifyPort)
	l, err := net.Listen("tcp", wsAddr)
	if err != nil {
		return err
	}
	log.Info("Listening for websockify connections on ", wsAddr)
	return srvr.ServeWebsockify(l)
}

func doListFeatures(authTypes []auth.Type, encTypes []encodings.Encoding, evTypes []events.Event) {
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

	w.Write([]byte("\nThe following features are available\n\n"))

	lformat := "%s\t(enabled)\n"

	fmt.Fprintln(w, "Security Types")
	fmt.Fprintln(w, "--------------")
	for _, sec := range authTypes {
		fmt.Fprintf(w, lformat, reflect.TypeOf(sec).Elem().Name())
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Encodings")
	fmt.Fprintln(w, "---------")
	for _, enc := range encTypes {
		fmt.Fprintf(w, lformat, reflect.TypeOf(enc).Elem().Name())
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Events")
	fmt.Fprintln(w, "------")
	for _, ev := range evTypes {
		fmt.Fprintf(w, lformat, reflect.TypeOf(ev).Elem().Name())
	}

	w.Flush()
	fmt.Println(buf.String())
}

func configureFeatures(args []string) ([]auth.Type, []encodings.Encoding, []events.Event) {
	return configureAuthTypes(auth.GetDefaults(), args),
		configureEncodings(encodings.GetDefaults(), args),
		configureEvents(events.GetDefaults(), args)
}

func authIsEnabled(tt []auth.Type, name string) bool {
	for _, t := range tt {
		if reflect.TypeOf(t).Elem().Name() == name {
			return true
		}
	}
	return false
}

func configureAuthTypes(tt []auth.Type, args []string) []auth.Type {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			tt = removeAuthType(tt, strings.TrimPrefix(arg, "-"))
		}
	}
	return tt
}

func configureEncodings(tt []encodings.Encoding, args []string) []encodings.Encoding {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			tt = removeEncoding(tt, strings.TrimPrefix(arg, "-"))
		}
	}
	return tt
}

func configureEvents(tt []events.Event, args []string) []events.Event {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			tt = removeEvent(tt, strings.TrimPrefix(arg, "-"))
		}
	}
	return tt
}

func removeAuthType(tt []auth.Type, name string) []auth.Type {
	newTT := make([]auth.Type, 0)
	for _, present := range tt {
		if reflect.TypeOf(present).Elem().Name() != name {
			newTT = append(newTT, present)
		}
	}
	return newTT
}

func removeEncoding(tt []encodings.Encoding, name string) []encodings.Encoding {
	newTT := make([]encodings.Encoding, 0)
	for _, present := range tt {
		if reflect.TypeOf(present).Elem().Name() != name {
			newTT = append(newTT, present)
		}
	}
	return newTT
}

func removeEvent(tt []events.Event, name string) []events.Event {
	newTT := make([]events.Event, 0)
	for _, present := range tt {
		if reflect.TypeOf(present).Elem().Name() != name {
			newTT = append(newTT, present)
		}
	}
	return newTT
}
