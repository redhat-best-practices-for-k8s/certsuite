package datautil

func IsMapSubset[K, V comparable](m, s map[K]V) bool {
	if len(s) > len(m) {
		return false
	}
	for ks, vs := range s {
		if vm, found := m[ks]; !found || vm != vs {
			return false
		}
	}
	return true
}
