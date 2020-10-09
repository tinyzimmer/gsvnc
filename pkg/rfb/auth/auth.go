package auth

import (
	"reflect"

	"github.com/tinyzimmer/gsvnc/pkg/buffer"
)

// Type represents an authentication type.
type Type interface {
	Code() uint8
	Negotiate(wr *buffer.ReadWriter) error
}

// EnabledAuthTypes is the currently enabled list of auth types. It can be mutated
// by command line optionss.
var EnabledAuthTypes = []Type{
	&None{},
	&VNCAuth{},
	&TightSecurity{},
}

// IsSupported returns true if the given auth type is supported.
func IsSupported(code uint8) bool {
	for _, t := range EnabledAuthTypes {
		if t.Code() == code {
			return true
		}
	}
	return false
}

// TightIsEnabled returns true if TightSecurity is enabled. This is used to determine if
// capabilities being mutated by the user also need to be updated here.
func TightIsEnabled() bool {
	t := &TightSecurity{}
	for _, a := range EnabledAuthTypes {
		if a.Code() == t.Code() {
			return true
		}
	}
	return false
}

// DisableAuth removes the given auth from the list of EnabledAuthTypes.
func DisableAuth(auth Type) {
	EnabledAuthTypes = remove(EnabledAuthTypes, auth)
}

// DisableTightAuth removes the given auth from the TightSecurity auth types.
func DisableTightAuth(code int32) {
	TightAuthCapabilities = removeCap(TightAuthCapabilities, code)
}

// DisableTightEncoding removes the given encoding from the TightSecurity encoding types.
func DisableTightEncoding(code int32) {
	TightEncodingCapabilities = removeCap(TightEncodingCapabilities, code)
}

func removeCap(cc []Capability, code int32) []Capability {
	newCaps := make([]Capability, 0)
	for _, enabled := range cc {
		if enabled.Code != code {
			newCaps = append(newCaps, enabled)
		}
	}
	return newCaps
}

func remove(tt []Type, t Type) []Type {
	newTypes := make([]Type, 0)
	for _, enabled := range tt {
		if reflect.TypeOf(enabled).Elem().Name() != reflect.TypeOf(t).Elem().Name() {
			newTypes = append(newTypes, enabled)
		}
	}
	return newTypes
}

// GetAuth returns the auth handler for the given code.
func GetAuth(code uint8) Type {
	for _, t := range EnabledAuthTypes {
		if t.Code() == code {
			return t
		}
	}
	return nil
}

// GetNone is a convenience wrapper for retrieving the noauth handler.
func GetNone() Type { return &None{} }
