package rfb

import (
	"net"
	"net/http"
	"reflect"
	"time"

	"golang.org/x/net/websocket"

	"github.com/tinyzimmer/gsvnc/pkg/display/providers"
	"github.com/tinyzimmer/gsvnc/pkg/internal/log"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/auth"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/encodings"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/events"
)

// ServerOpts represents options that can be used to configure a new RFB server.
type ServerOpts struct {
	DisplayProvider  providers.Provider
	Width, Height    int
	ServerPassword   string
	EnabledEncodings []encodings.Encoding
	EnabledAuthTypes []auth.Type
	EnabledEvents    []events.Event
}

// NewServer creates a new RFB server with an initial width and height.
func NewServer(opts *ServerOpts) *Server {
	server := &Server{
		displayProvider:  opts.DisplayProvider,
		width:            opts.Width,
		height:           opts.Height,
		serverPassword:   opts.ServerPassword,
		enabledEncodings: opts.EnabledEncodings,
		enabledAuthTypes: opts.EnabledAuthTypes,
		enabledEvents:    opts.EnabledEvents,
	}

	// Configure default events if any are empty
	if len(opts.EnabledEncodings) == 0 {
		server.enabledEncodings = encodings.GetDefaults()
	}
	if len(opts.EnabledAuthTypes) == 0 {
		server.enabledAuthTypes = auth.GetDefaults()
	}
	if len(opts.EnabledEvents) == 0 {
		server.enabledEvents = events.GetDefaults()
	}

	// Configure tight if enabled
	if server.TightIsEnabled() {
		iface := server.GetAuthByName("TightSecurity")
		tight := iface.(*auth.TightSecurity)
		tight.AuthGetter = server.GetAuth
		// TODO: Configure capabilities
	}

	return server
}

// Server represents an RFB server. A channel is exposed for handling incoming client
// connections.
type Server struct {
	width, height    int
	serverPassword   string
	displayProvider  providers.Provider
	enabledEncodings []encodings.Encoding
	enabledAuthTypes []auth.Type
	enabledEvents    []events.Event
}

// Serve binds the RFB server to the given listener and starts serving connections.
func (s *Server) Serve(ln net.Listener) error {
	for {

		// Accept a new connection
		c, err := ln.Accept()
		if err != nil {
			return err
		}

		log.Info("New client connection from ", c.RemoteAddr().String())

		// Create a new client connection
		conn := s.newConn(c)

		// Do the rfb handshake
		if err := conn.doHandshake(); err != nil {
			log.Error("Error during server-client handshake: ", err.Error())
			conn.c.Close()
			continue
		}

		// handle events
		go conn.serve()
	}
}

// ServeWebsockify will serve websockify connections on the given host and port.
func (s *Server) ServeWebsockify(ln net.Listener) error {
	srvr := &http.Server{
		Addr:        ln.Addr().String(),
		ReadTimeout: time.Second * 300, WriteTimeout: time.Second * 300,
		Handler: &websocket.Server{
			Handshake: func(cfg *websocket.Config, r *http.Request) error { return nil },
			Handler: func(wsconn *websocket.Conn) {
				log.Info("New websocket client connection from ", wsconn.Request().RemoteAddr)
				wsconn.PayloadType = websocket.BinaryFrame
				// Create a new client connection
				conn := s.newConn(wsconn)

				// Do the rfb handshake
				if err := conn.doHandshake(); err != nil {
					log.Error("Error during server-client handshake: ", err.Error())
					conn.c.Close()
					return
				}

				// handle events
				conn.serve()
			},
		},
	}
	return srvr.Serve(ln)
}

// AuthIsSupported returns true if the given auth type is supported.
func (s *Server) AuthIsSupported(code uint8) bool {
	for _, t := range s.enabledAuthTypes {
		if t.Code() == code {
			return true
		}
	}
	return false
}

// VNCAuthIsEnabled returns true if VNCAuth is enabled on the server. This is used to signal
// the need to generate (or, in the future, read in) the server password.
func (s *Server) VNCAuthIsEnabled() bool {
	t := &auth.VNCAuth{}
	for _, a := range s.enabledAuthTypes {
		if a.Code() == t.Code() {
			return true
		}
	}
	return false
}

// TightIsEnabled returns true if TightSecurity is enabled. This is used to determine if
// capabilities being mutated by the user also need to be updated here.
func (s *Server) TightIsEnabled() bool {
	t := &auth.TightSecurity{}
	for _, a := range s.enabledAuthTypes {
		if a.Code() == t.Code() {
			return true
		}
	}
	return false
}

// GetAuth returns the auth handler for the given code.
func (s *Server) GetAuth(code uint8) auth.Type {
	for _, t := range s.enabledAuthTypes {
		if t.Code() == code {
			return t
		}
	}
	return nil
}

// GetAuthByName returns the auth interface by the given name.
func (s *Server) GetAuthByName(name string) auth.Type {
	for _, t := range s.enabledAuthTypes {
		if reflect.TypeOf(t).Elem().Name() == name {
			return t
		}
	}
	return nil
}

// GetEventHandlerMap returns a map that can be used for handling events on an
// rfb connection.
func (s *Server) GetEventHandlerMap() map[uint8]events.Event {
	out := make(map[uint8]events.Event)
	for _, ev := range s.enabledEvents {
		out[ev.Code()] = ev
	}
	return out
}

// GetEncoding will iterate the requested encodings and return the best match
// that can be served. If none of the requested encodings are supported (should
// never happen as at least RAW is required by RFC) this function returns nil.
func (s *Server) GetEncoding(encs []int32) encodings.Encoding {
	for _, e := range encs {
		for _, supported := range s.enabledEncodings {
			if e == supported.Code() {
				log.Debugf("Using %s encoding", reflect.TypeOf(supported).Elem().Name())
				return supported
			}
		}
	}
	return nil
}
