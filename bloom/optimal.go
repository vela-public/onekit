package bloom

import "math"

// optimalSize calculates the optimal Size of the Bitset
func optimalSize(n int, p float64) int {
	m := -float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)
	return int(math.Ceil(m))
}

// optimalHashFunctions calculates the optimal number of hash functions
func optimalHashFunctions(m, n int) int {
	k := float64(m) / float64(n) * math.Ln2
	return int(math.Ceil(k))
}
