package ngram

import "github.com/vela-public/onekit/strkit/internal/mathutil"

// Count returns the n-gram count of the specified size for the
// provided term. An n-gram size of 1 is used if the provided size is
// less than or equal to 0.
func Count(runes []rune, size int) int {
	return mathutil.Max(len(runes)-(mathutil.Max(size, 1)-1), 0)
}

// Slice returns all the n-grams of the specified size for the provided term.
// The n-grams in the output slice are in the order in which they occur in the
// input term. An n-gram size of 1 is used if the provided size is less than
// or equal to 0.
func Slice(runes []rune, size int) []string {
	// Use an n-gram size of 1 if the provided size is invalid.
	size = mathutil.Max(size, 1)

	// Check if term length is too small.
	lenRunes := len(runes)
	if lenRunes == 0 || lenRunes < size {
		return nil
	}

	// Generate n-gram slice.
	limit := lenRunes - (size - 1)
	ngrams := make([]string, limit)

	for i, j := 0, 0; i < limit; i++ {
		ngrams[j] = string(runes[i : i+size])
		j++
	}

	return ngrams
}

// Map returns a map of all n-grams of the specified size for the provided
// term, along with their frequency. The function also returns the total
// number of n-grams, which is the sum of all the values in the output map.
// An n-gram size of 1 is used if the provided size is less than or equal to 0.
func Map(runes []rune, size int) (map[string]int, int) {
	// Use an n-gram size of 1 if the provided size is invalid.
	size = mathutil.Max(size, 1)

	// Check if term length is too small.
	lenRunes := len(runes)
	if lenRunes == 0 || lenRunes < size {
		return map[string]int{}, 0
	}

	// Generate n-gram map.
	limit := lenRunes - (size - 1)
	ngrams := make(map[string]int, limit)

	var ngramCount int
	for i := 0; i < limit; i++ {
		ngram := string(runes[i : i+size])
		count, _ := ngrams[ngram]
		ngrams[ngram] = count + 1
		ngramCount++
	}

	return ngrams, ngramCount
}

// Intersection returns a map of the n-grams of the specified size found
// in both terms, along with their frequency. The function also returns the
// number of common n-grams (the sum of all the values in the output map),
// the total number of n-grams in the first term and the total number of
// n-grams in the second term. An n-gram size of 1 is used if the provided
// size is less than or equal to 0.
func Intersection(a, b []rune, size int) (map[string]int, int, int, int) {
	// Use an n-gram size of 1 if the provided size is invalid.
	size = mathutil.Max(size, 1)

	// Compute the n-grams of the first term.
	ngramsA, totalA := Map(a, size)

	// Calculate n-gram intersection with the second term.
	limit := len(b) - (size - 1)
	commonNgrams := make(map[string]int, mathutil.Max(limit, 0))

	var totalB, intersection int
	for i := 0; i < limit; i++ {
		ngram := string(b[i : i+size])
		totalB++

		if count, ok := ngramsA[ngram]; ok && count > 0 {
			// Decrease frequency of n-gram found in the first term each time
			// a successful match is found.
			intersection++
			ngramsA[ngram] = count - 1

			// Update common n-grams map with the matched n-gram and its
			// frequency.
			count, _ = commonNgrams[ngram]
			commonNgrams[ngram] = count + 1
		}
	}

	return commonNgrams, intersection, totalA, totalB
}
