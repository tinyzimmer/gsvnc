package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/tinyzimmer/gsvnc/pkg/encodings"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/auth"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/events"
)

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

			// Auth types
			for _, sec := range auth.EnabledAuthTypes {
				if reflect.TypeOf(sec).Elem().Name() == featName {
					if auth.TightIsEnabled() {
						auth.DisableTightAuth(int32(sec.Code()))
					}
					auth.DisableAuth(sec)
					continue ArgLoop
				}
			}

			// Encodings
			for _, enc := range encodings.EnabledEncodings {
				if reflect.TypeOf(enc).Elem().Name() == featName {
					if auth.TightIsEnabled() {
						auth.DisableTightAuth(int32(enc.Code()))
					}
					encodings.DisableEncoding(enc)
					continue ArgLoop
				}
			}

			// Event handlers
			for _, ev := range events.EnabledEvents {
				if reflect.TypeOf(ev).Elem().Name() == featName {
					events.DisableEvent(ev)
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
