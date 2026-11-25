package game

import (
	"log"
	"syscall/js"
)

// primaryTouchControls reports whether the browser's primary pointer is coarse.
func primaryTouchControls() bool {
	matchMedia := js.Global().Get("matchMedia")
	if matchMedia.Type() != js.TypeFunction {
		log.Print("controls: browser does not support matchMedia; showing keyboard instructions")
		return false
	}
	return js.Global().Call("matchMedia", "(pointer: coarse)").Get("matches").Bool()
}
