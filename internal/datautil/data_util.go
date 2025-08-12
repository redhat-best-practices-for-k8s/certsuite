package datautil

// IsMapSubset reports whether the first map is a subset of the second.
//
// It iterates over each key-value pair in the first map and checks that the same
// key exists in the second map with an identical value. If all entries match,
// it returns true; otherwise it returns false. The function does not modify
// either input map.
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
