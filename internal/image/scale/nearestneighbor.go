package scale

import (
	"image"
	"image/color"
	"math"
)

func NearestNeighbor(src image.Image, w, h uint) image.Image {
	b := src.Bounds()
	fx := float64(w) / float64(b.Dx())
	fy := float64(h) / float64(b.Dy())
	dst := image.NewNRGBA(image.Rect(0, 0, int(w), int(h)))
	for y := range int(h) {
		sy := clamp(int(math.Round(float64(y)/fy)), b.Min.Y, b.Max.Y-1)
		for x := range int(w) {
			sx := clamp(int(math.Round(float64(x)/fx)), b.Min.X, b.Max.X-1)
			c := color.NRGBAModel.Convert(src.At(sx, sy)).(color.NRGBA)
			dst.Set(x, y, c)
		}
	}
	return dst
}
