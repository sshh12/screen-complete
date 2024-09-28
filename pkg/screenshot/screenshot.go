package screenshot

import (
	"image"

	"github.com/go-vgo/robotgo"
)

func CaptureBounds(x1, y1, x2, y2 int) *image.RGBA {
	// Calculate width and height from mouse coordinates
	w := x2 - x1
	h := y2 - y1

	// Ensure positive width and height
	if w < 0 {
		x1, x2 = x2, x1
		w = -w
	}
	if h < 0 {
		y1, y2 = y2, y1
		h = -h
	}

	bitmap := robotgo.CaptureScreen(x1, y1, w, h)
	defer robotgo.FreeBitmap(bitmap)

	// Convert bitmap to image.RGBA
	img := robotgo.ToRGBA(bitmap)
	return img
}