package game

import "testing"

// TestSwipeDirection checks the threshold and the dominant-axis rule of swipe detection.
func TestSwipeDirection(t *testing.T) {
	cases := []struct {
		dx, dy int
		want   Direction
	}{
		{0, 0, DirectionNone},
		{swipeThreshold - 1, 0, DirectionNone},
		{0, -(swipeThreshold - 1), DirectionNone},
		{swipeThreshold, 0, DirectionRight},
		{-swipeThreshold, 0, DirectionLeft},
		{0, -swipeThreshold, DirectionUp},
		{0, swipeThreshold, DirectionDown},
		{40, -30, DirectionRight},
		{-30, 40, DirectionDown},
		{swipeThreshold, swipeThreshold, DirectionRight},
		{swipeThreshold - 1, swipeThreshold - 1, DirectionNone},
	}
	for _, c := range cases {
		if got := swipeDirection(c.dx, c.dy); got != c.want {
			t.Fatalf("swipeDirection(%d, %d) = %v, want %v", c.dx, c.dy, got, c.want)
		}
	}
}

// TestTouchSwipeSteersWithoutLifting checks that a drag reports a direction once per
// threshold crossing and that a turn in the same drag reports the new direction.
func TestTouchSwipeSteersWithoutLifting(t *testing.T) {
	touch := newTouchInput()
	touch.Press(1, 100, 100)
	if got := touch.Move(1, 100+swipeThreshold-1, 100); got != DirectionNone {
		t.Fatalf("short drag should report no direction, got %v", got)
	}
	if got := touch.Move(1, 100+swipeThreshold, 100); got != DirectionRight {
		t.Fatalf("drag past the threshold should report right, got %v", got)
	}
	if got := touch.Move(1, 100+swipeThreshold+5, 100); got != DirectionNone {
		t.Fatalf("holding after a swipe should report nothing until the next threshold, got %v", got)
	}
	if got := touch.Move(1, 100+swipeThreshold+5, 100-swipeThreshold); got != DirectionUp {
		t.Fatalf("turning the drag upward should report up, got %v", got)
	}
	if gesture, _, _ := touch.Release(1); gesture != TouchGestureNone {
		t.Fatalf("lifting after a swipe should not be a tap, got %v", gesture)
	}
	if got := touch.Move(1, 0, 0); got != DirectionNone {
		t.Fatalf("a lifted finger should report no direction, got %v", got)
	}
}

// TestTouchTaps checks single and two-finger taps, the tap position, and that a third
// finger or a swipe cancels the tap.
func TestTouchTaps(t *testing.T) {
	touch := newTouchInput()
	touch.Press(1, 50, 20)
	touch.Move(1, 55, 24)
	gesture, x, y := touch.Release(1)
	if gesture != TouchGestureTap || x != 50 || y != 20 {
		t.Fatalf("single tap = %v at (%d, %d), want tap at (50, 20)", gesture, x, y)
	}

	touch.Press(1, 50, 300)
	touch.Press(2, 200, 300)
	if gesture, _, _ := touch.Release(1); gesture != TouchGestureNone {
		t.Fatalf("lifting the first of two fingers should report nothing, got %v", gesture)
	}
	if gesture, _, _ := touch.Release(2); gesture != TouchGestureTwoFingerTap {
		t.Fatalf("lifting the last of two fingers should be a two-finger tap, got %v", gesture)
	}

	touch.Press(1, 50, 300)
	touch.Press(2, 200, 300)
	touch.Release(2)
	touch.Press(3, 300, 300)
	touch.Release(1)
	if gesture, _, _ := touch.Release(3); gesture != TouchGestureNone {
		t.Fatalf("three fingers in one gesture should not be a tap, got %v", gesture)
	}

	touch.Press(1, 50, 300)
	touch.Press(2, 200, 300)
	touch.Move(2, 200+swipeThreshold, 300)
	touch.Release(2)
	if gesture, _, _ := touch.Release(1); gesture != TouchGestureNone {
		t.Fatalf("a gesture with a swipe should not be a tap, got %v", gesture)
	}

	touch.Press(1, 10, 10)
	if gesture, _, _ := touch.Release(1); gesture != TouchGestureTap {
		t.Fatalf("a new gesture after a cancelled one should be a tap again, got %v", gesture)
	}
	if gesture, _, _ := touch.Release(9); gesture != TouchGestureNone {
		t.Fatalf("releasing an unknown finger should report nothing, got %v", gesture)
	}
}

// TestTouchPauseZone checks that the pause zone covers exactly the top HUD rows.
func TestTouchPauseZone(t *testing.T) {
	if !touchInPauseZone(0) || !touchInPauseZone(touchPauseRows*CellSize-1) {
		t.Fatal("the top HUD rows should be in the pause zone")
	}
	if touchInPauseZone(touchPauseRows * CellSize) {
		t.Fatal("the first maze row should not be in the pause zone")
	}
}
