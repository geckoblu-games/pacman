//go:build !js

package game

import "runtime"

// primaryTouchControls selects touch instructions for native mobile hosts.
func primaryTouchControls() bool {
	return runtime.GOOS == "android" || runtime.GOOS == "ios"
}
