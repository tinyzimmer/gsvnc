package auth

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/internal/util"
)

// Capability represents a TightSecurity capability.
type Capability struct {
	Code              int32
	Vendor, Signature string
}

// TightTunnelCapabilities represents TightSecurity tunnel capabilities.
var TightTunnelCapabilities = []Capability{
	{Code: 0, Vendor: "TGHT", Signature: "NOTUNNEL"},
}

// TightAuthCapabilities represents TightSecurity auth capabilities.
var TightAuthCapabilities = []Capability{
	{Code: 1, Vendor: "STDV", Signature: "NOAUTH__"},
	{Code: 2, Vendor: "STDV", Signature: "VNCAUTH_"},
}

// TightServerMessages represents supported tight server messages.
var TightServerMessages = []Capability{}

// TightClientMessages represents supported tight client messages.
var TightClientMessages = []Capability{}

// TightEncodingCapabilities represents TightSecurity encoding capabilities.
var TightEncodingCapabilities = []Capability{
	{Code: 0, Vendor: "STDV", Signature: "RAW_____"},
	{Code: 1, Vendor: "STDV", Signature: "COPYRECT"},
	{Code: 7, Vendor: "TGHT", Signature: "TIGHT___"},
}

// TightSecurity implements Tight security.
// https://github.com/rfbproto/rfbproto/blob/master/rfbproto.rst#tight-security-type
type TightSecurity struct{}

// Code returns the code.
func (t *TightSecurity) Code() uint8 { return 16 }

// Negotiate will negotiate tight security.
func (t *TightSecurity) Negotiate(rw *buffer.ReadWriter) error {
	if err := negotiateTightTunnel(rw); err != nil {
		return err
	}
	return negotiateTightAuth(rw)
}

// ExtendServerInit signals to the rfb server that we extend the ServerInit message.
func (t *TightSecurity) ExtendServerInit(buf io.Writer) {
	util.Write(buf, uint16(len(TightServerMessages)))
	util.Write(buf, uint16(len(TightClientMessages)))
	util.Write(buf, uint16(len(TightEncodingCapabilities)))
	util.Write(buf, uint8(0)) // padding
	util.Write(buf, uint8(0)) // padding
	for _, cap := range TightServerMessages {
		util.PackStruct(buf, &cap)
	}
	for _, cap := range TightClientMessages {
		util.PackStruct(buf, &cap)
	}
	for _, cap := range TightEncodingCapabilities {
		util.PackStruct(buf, &cap)
	}
}

func negotiateTightTunnel(rw *buffer.ReadWriter) error {
	// Write the supported tunnel capabilities to the client
	buf := new(bytes.Buffer)
	util.Write(buf, uint32(len(TightTunnelCapabilities)))
	for _, cap := range TightTunnelCapabilities {
		util.PackStruct(buf, &cap)
	}
	rw.Dispatch(buf.Bytes())

	// get the desired tunnel type from the client
	var tun int32
	rw.Read(&tun)

	// We only support no tunneling for now, client should know
	// better
	if tun != 0 {
		return fmt.Errorf("client requested unsupported tunnel type: %d", tun)
	}

	return nil
}

func negotiateTightAuth(rw *buffer.ReadWriter) error {
	buf := new(bytes.Buffer)
	util.Write(buf, uint32(len(TightAuthCapabilities)))
	for _, cap := range TightAuthCapabilities {
		util.PackStruct(buf, &cap)
	}
	rw.Dispatch(buf.Bytes())

	// get the desired auth type, should be none
	var auth int32
	rw.Read(&auth)

	if !IsSupported(uint8(auth)) {
		return fmt.Errorf("client requested unsupported tight auth type: %d", auth)
	}

	authType := GetAuth(uint8(auth))
	return authType.Negotiate(rw)
}
