package set

type Set[T comparable] struct {
	m map[T]struct{}
}

func New[T comparable](v ...T) *Set[T] {
	if len(v) != 0 {
		m := make(map[T]struct{}, len(v))
		for _, v := range v {
			m[v] = struct{}{}
		}
		return &Set[T]{m}
	}
	return &Set[T]{make(map[T]struct{})}
}

func (s Set[T]) Len() int {
	return len(s.m)
}

func (s Set[T]) Add(v T) {
	s.m[v] = struct{}{}
}

func (s Set[T]) Has(v T) bool {
	_, ok := s.m[v]
	return ok
}

func (s Set[T]) Values() []T {
	vv := make([]T, 0, len(s.m))
	for v := range s.m {
		vv = append(vv, v)
	}
	return vv
}
