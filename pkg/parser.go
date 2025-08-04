package pkg

func mergeMap[TK comparable, TV any](m1, m2 map[TK]TV) map[TK]TV {
	m := make(map[TK]TV)
	for tk, tv := range m1 {
		m[tk] = tv
	}
	for tk, tv := range m2 {
		m[tk] = tv
	}
	return m
}
