package scale

import "cmp"

func clamp[T cmp.Ordered](x, a, b T) T {
	if x < a {
		return a
	}
	if x > b {
		return b
	}
	return x
}
