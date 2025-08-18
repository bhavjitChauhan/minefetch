package quant

import (
	"image/color"
	"sort"
)

type colorChannel int

const (
	red = iota
	blue
	green
)

type bucket struct {
	colors     []color.NRGBA
	colorRange color.NRGBA
	maxRangeCh colorChannel
}

func newBucket(colors []color.NRGBA) *bucket {
	b := &bucket{colors: colors}
	b.findRange()
	return b
}

func (b bucket) cut() (*bucket, *bucket) {
	sort.Sort(b)
	b0 := newBucket(b.colors[:len(b.colors)/2])
	b1 := newBucket(b.colors[len(b.colors)/2:])
	return b0, b1
}

func (b bucket) mean() color.NRGBA {
	n := len(b.colors)
	if n == 1 {
		return b.colors[0]
	}
	var sumR, sumG, sumB uint
	for _, c := range b.colors {
		sumR += uint(c.R)
		sumB += uint(c.B)
		sumG += uint(c.G)
	}
	sumR /= uint(n)
	sumG /= uint(n)
	sumB /= uint(n)
	return color.NRGBA{uint8(sumR), uint8(sumG), uint8(sumB), 255}
}

func (b *bucket) findRange() {
	min := b.colors[0]
	max := b.colors[0]
	for _, c := range b.colors {
		if c.R < min.R {
			min.R = c.R
		} else if c.R > max.R {
			max.R = c.R
		}
		if c.G < min.G {
			min.G = c.G
		} else if c.G > max.G {
			max.G = c.G
		}
		if c.B < min.B {
			min.B = c.B
		} else if c.B > max.B {
			max.B = c.B
		}
	}
	b.colorRange.R = max.R - min.R
	b.colorRange.G = max.G - min.G
	b.colorRange.B = max.B - min.B
	switch {
	case b.colorRange.R >= b.colorRange.G && b.colorRange.R >= b.colorRange.B:
		b.maxRangeCh = red
	case b.colorRange.G >= b.colorRange.R && b.colorRange.G >= b.colorRange.B:
		b.maxRangeCh = green
	default:
		b.maxRangeCh = blue
	}
}

func (b bucket) Len() int {
	return len(b.colors)
}

func (b bucket) Swap(i, j int) {
	b.colors[i], b.colors[j] = b.colors[j], b.colors[i]
}

func (b bucket) Less(i, j int) bool {
	switch b.maxRangeCh {
	case red:
		return b.colors[i].R < b.colors[j].R
	case green:
		return b.colors[i].G < b.colors[j].G
	case blue:
		return b.colors[i].B < b.colors[j].B
	}
	panic("invalid color channel")
}

type buckets []*bucket

func (bb buckets) Len() int {
	return len(bb)
}

func (bb buckets) Swap(i, j int) {
	bb[i], bb[j] = bb[j], bb[i]
}

func (bb buckets) Less(i, j int) bool {
	maxI := max(bb[i].colorRange.R, bb[i].colorRange.G, bb[i].colorRange.B)
	maxJ := max(bb[j].colorRange.R, bb[j].colorRange.G, bb[j].colorRange.B)
	return maxI > maxJ
}

func (bb *buckets) Push(x any) {
	*bb = append(*bb, x.(*bucket))
}

func (bb *buckets) Pop() any {
	old := *bb
	i := len(old) - 1
	b := old[i]
	old[i] = nil
	*bb = old[:i]
	return b
}
