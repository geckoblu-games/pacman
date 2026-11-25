package game

const (
	// swipeThreshold is the travel in logical pixels along the dominant axis that turns a
	// drag into a swipe; one cell keeps taps distinguishable from short swipes.
	swipeThreshold = CellSize
	// touchPauseRows counts the top grid rows forming the pause tap zone, the HUD above
	// the maze.
	touchPauseRows = 3
)

// TouchGesture identifies the tap gesture reported when the last finger lifts.
type TouchGesture int

const (
	// TouchGestureNone means the lifted touches were a swipe or more than two fingers.
	TouchGestureNone TouchGesture = iota
	// TouchGestureTap is a single finger pressed and lifted without swiping.
	TouchGestureTap
	// TouchGestureTwoFingerTap is two fingers pressed and lifted without swiping.
	TouchGestureTwoFingerTap
)

// touchPoint tracks one held finger by the logical position its next swipe is measured from.
type touchPoint struct {
	// originX and originY are where the finger landed or where its last swipe completed.
	originX, originY int
}

// TouchInput converts raw touch positions into Pac-Man directions and tap gestures,
// replacing the keyboard on touch screens: a swipe anywhere requests a direction, a tap
// starts a match or, on the top HUD during a match, pauses, and a two-finger tap mutes.
// It is renderer-independent; ScenePlay feeds it Ebitengine's touch IDs and logical
// positions every tick.
type TouchInput struct {
	// points holds the fingers currently on the screen by ID.
	points map[int]touchPoint
	// fingers counts the fingers that have landed since the current gesture started.
	fingers int
	// swiped records whether the current gesture produced a swipe, which disqualifies it
	// as a tap.
	swiped bool
	// lastX and lastY are the landing position of the most recently lifted finger.
	lastX, lastY int
}

// newTouchInput returns a tracker with no fingers on the screen.
func newTouchInput() *TouchInput {
	return &TouchInput{points: map[int]touchPoint{}}
}

// Press registers a finger landing at a logical screen position. A gesture starts with
// the first finger and ends when the last one lifts.
func (touch *TouchInput) Press(id, x, y int) {
	if len(touch.points) == 0 {
		touch.fingers = 0
		touch.swiped = false
	}
	touch.points[id] = touchPoint{originX: x, originY: y}
	touch.fingers++
}

// Move reports a held finger's current position and returns the swiped direction once
// it has travelled swipeThreshold from its origin along the dominant axis, moving the
// origin to the current position so that a continued drag can steer again without
// lifting. It returns DirectionNone for unknown fingers and shorter travel.
func (touch *TouchInput) Move(id, x, y int) Direction {
	point, ok := touch.points[id]
	if !ok {
		return DirectionNone
	}
	direction := swipeDirection(x-point.originX, y-point.originY)
	if direction == DirectionNone {
		return DirectionNone
	}
	touch.points[id] = touchPoint{originX: x, originY: y}
	touch.swiped = true
	return direction
}

// Release removes a lifted finger. When it is the last finger of the gesture, Release
// returns the tap gesture and the lifted finger's landing position; otherwise, and for
// gestures that swiped or landed more than two fingers, it returns TouchGestureNone.
func (touch *TouchInput) Release(id int) (TouchGesture, int, int) {
	point, ok := touch.points[id]
	if !ok {
		return TouchGestureNone, 0, 0
	}
	delete(touch.points, id)
	touch.lastX, touch.lastY = point.originX, point.originY
	if len(touch.points) > 0 || touch.swiped {
		return TouchGestureNone, touch.lastX, touch.lastY
	}
	switch touch.fingers {
	case 1:
		return TouchGestureTap, touch.lastX, touch.lastY
	case 2:
		return TouchGestureTwoFingerTap, touch.lastX, touch.lastY
	default:
		return TouchGestureNone, touch.lastX, touch.lastY
	}
}

// swipeDirection maps a drag displacement in logical pixels to the cardinal direction of
// its dominant axis, or DirectionNone while that axis is short of swipeThreshold. Equal
// displacements favor the horizontal axis.
func swipeDirection(dx, dy int) Direction {
	absX, absY := dx, dy
	if absX < 0 {
		absX = -absX
	}
	if absY < 0 {
		absY = -absY
	}
	switch {
	case absX >= absY && absX >= swipeThreshold:
		if dx < 0 {
			return DirectionLeft
		}
		return DirectionRight
	case absY > absX && absY >= swipeThreshold:
		if dy < 0 {
			return DirectionUp
		}
		return DirectionDown
	default:
		return DirectionNone
	}
}

// touchInPauseZone reports whether a logical position lies on the top HUD rows that pause
// the match when tapped.
func touchInPauseZone(y int) bool {
	return y < touchPauseRows*CellSize
}
