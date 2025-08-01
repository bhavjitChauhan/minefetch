package scale

import (
	"image"
	"image/color"
	"math"
)

// Lanczos scales the image src by a factor of f using three-lobed [Lanczos resampling].
// This function is hard-coded to produce the best results for downscaling, i.e. f < 1.
// Upscaling will likely produce undesired artifacts.
//
// Sinc interpolation is viewed as theoretically the best method to scale images.
// However, the sinc function is unbounded and doesn't drop off.
// Lanczos resampling uses a windowed sinc function as its kernel.
//
// See [Jeff Boody's explainer] for a great resource on Lanczos resampling.
//
// [Lanczos resampling]: https://en.wikipedia.org/wiki/Lanczos_resampling
// [Jeff Boody's explainer]: https://github.com/jeffboody/Lanczos
func Lanczos(src image.Image, f float64) image.Image {
	const a = 3
	sb := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, int(math.Round(float64(sb.Dx())*f)), int(math.Round(float64(sb.Dy())*f))))
	db := dst.Bounds()
	fa := math.Round(a / f)
	for y := range db.Dy() {
		for x := range db.Dx() {
			var acc struct {
				r, g, b, a float64
			}
			sx := (float64(x)+.5)/f - .5
			sy := (float64(y)+.5)/f - .5
			w := 1 / weight(a, sx) * weight(a, sy) * f * f
			for j := -fa + 1; j < fa; j++ {
				lj := lanczos(uint(fa), (j-sy+math.Floor(sy))*f)
				for i := -fa + 1; i < fa; i++ {
					sxi := clamp(int(math.Floor(sx)+i), 0, sb.Dx()-1)
					syj := clamp(int(math.Floor(sy)+j), 0, sb.Dy()-1)
					wl := lanczos(uint(fa), (i-sx+math.Floor(sx))*f) * lj * w
					c := color.NRGBAModel.Convert(src.At(sxi, syj)).(color.NRGBA)
					acc.r += wl * float64(c.R)
					acc.g += wl * float64(c.G)
					acc.b += wl * float64(c.B)
					acc.a += wl * float64(c.A)
				}
			}
			acc.r = clamp(acc.r, 0, 255)
			acc.g = clamp(acc.g, 0, 255)
			acc.b = clamp(acc.b, 0, 255)
			acc.a = clamp(acc.a, 0, 255)
			dst.Set(x, y, color.NRGBA{uint8(acc.r), uint8(acc.g), uint8(acc.b), uint8(acc.a)})
		}
	}
	return dst
}

func weight(a uint, x float64) float64 {
	w := 0.0
	for i := -int(a) + 1; i < int(a); i++ {
		w += lanczos(a, float64(i)-x+math.Floor(x))
	}
	return w
}

func lanczos(a uint, x float64) float64 {
	if math.Abs(x) >= float64(a) {
		return 0
	}
	return sinc(x) * sinc(x/float64(a))
}

func sinc(x float64) float64 {
	if x == 0 {
		return 1
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}
