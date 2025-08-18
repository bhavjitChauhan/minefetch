package quant

import (
	"container/heap"
	"image"
	"image/color"
	"minefetch/internal/set"
)

// MedianCut is a color quantizer using the [median cut] algorithm.
//
// [median cut]: https://en.wikipedia.org/wiki/Median_cut
var MedianCut = medianCut{}

type medianCut struct{}

// Quantize implements the [Quantizer] interface for the median cut algorithm.
//
// [Quantizer]: https://pkg.go.dev/image/draw#Quantizer
func (medianCut) Quantize(p color.Palette, m image.Image) color.Palette {
	if len(p) == cap(p) {
		return p
	}
	bounds := m.Bounds()
	set := set.New[color.NRGBA]()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.NRGBAModel.Convert(m.At(x, y)).(color.NRGBA)
			// Ignore opacity
			// https://usage.imagemagick.org/quantize/#:~:text=Notice%20how%20the,to%20look%20like.
			c.A = 255
			set.Add(c)
		}
	}
	buckets := buckets{newBucket(set.Values())}
	heap.Init(&buckets)
	for len(buckets) < cap(p)-len(p) {
		b := heap.Pop(&buckets).(*bucket)
		if b.Len() == 1 {
			buckets = append(buckets, b)
			break
		}
		b0, b1 := b.cut()
		heap.Push(&buckets, b0)
		heap.Push(&buckets, b1)
	}
	for _, b := range buckets {
		p = append(p, b.mean())
	}
	return p
}
