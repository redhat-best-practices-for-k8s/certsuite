package datautil

// IsMapSubset Determines if one map contains all key-value pairs of another
//
// It compares two generic maps, returning true only when every entry in the
// second map exists identically in the first. The function first checks that
// the second map is not larger than the first for efficiency. It then iterates
// through each key-value pair, verifying presence and equality; if any mismatch
// occurs, it returns false.
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
