package auth

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/util"
)

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
	util.Write(buf, uint16(len(tightServerMessages)))
	util.Write(buf, uint16(len(tightClientMessages)))
	util.Write(buf, uint16(len(tightEncodingCapabilities)))
	util.Write(buf, uint8(0)) // padding
	util.Write(buf, uint8(0)) // padding
	for _, cap := range tightServerMessages {
		util.Write(buf, cap.code)
		util.Write(buf, []byte(cap.vendor))
		util.Write(buf, []byte(cap.signature))
	}
	for _, cap := range tightClientMessages {
		util.Write(buf, cap.code)
		util.Write(buf, []byte(cap.vendor))
		util.Write(buf, []byte(cap.signature))
	}
	for _, cap := range tightEncodingCapabilities {
		util.Write(buf, cap.code)
		util.Write(buf, []byte(cap.vendor))
		util.Write(buf, []byte(cap.signature))
	}
}

type capability struct {
	code              int32
	vendor, signature string
}

var tightTunnelCapabilities = []capability{
	{code: 0, vendor: "TGHT", signature: "NOTUNNEL"},
}

var tightAuthCapabilities = []capability{
	{code: 1, vendor: "STDV", signature: "NOAUTH__"},
}

var tightServerMessages = []capability{}
var tightClientMessages = []capability{}

// TODO: this would be altered by command line options technically
var tightEncodingCapabilities = []capability{
	{code: 0, vendor: "STDV", signature: "RAW_____"},
	{code: 1, vendor: "STDV", signature: "COPYRECT"},
	{code: 7, vendor: "TGHT", signature: "TIGHT___"},
}

func negotiateTightTunnel(rw *buffer.ReadWriter) error {
	// Write the supported tunnel capabilities to the client
	buf := new(bytes.Buffer)
	util.Write(buf, uint32(len(tightTunnelCapabilities)))
	for _, cap := range tightTunnelCapabilities {
		util.Write(buf, cap.code)
		util.Write(buf, []byte(cap.vendor))
		util.Write(buf, []byte(cap.signature))
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
	util.Write(buf, uint32(len(tightAuthCapabilities)))
	for _, cap := range tightAuthCapabilities {
		util.Write(buf, cap.code)
		util.Write(buf, []byte(cap.vendor))
		util.Write(buf, []byte(cap.signature))
	}
	rw.Dispatch(buf.Bytes())

	// get the desired auth type, should be none
	var auth int32
	rw.Read(&auth)

	// We only support no auth for now
	if auth != 1 {
		return fmt.Errorf("client requested unsupported tight auth type: %d", auth)
	}

	return nil
}
