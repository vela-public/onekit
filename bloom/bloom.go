package bloom

import (
	"github.com/spaolacci/murmur3"
	"github.com/vela-public/onekit/bitset"
)

type Filter struct {
	Bitset bitset.BitSet `json:"bitset"`
	Size   uint          `json:"size"`
	Hashes int           `json:"hashes"`
	Cnt    int           `json:"cnt"`
}

// New creates a new Filter with a given number of items and false positive rate
func New(numItems int, falsePositiveRate float64) *Filter {
	size := optimalSize(numItems, falsePositiveRate)
	hashes := optimalHashFunctions(size, numItems)
	return &Filter{
		Bitset: *bitset.New(uint(size)),
		Size:   uint(size),
		Hashes: hashes,
	}
}

func (bf *Filter) Sizeof() int {
	sz := 0
	sz = sz + 8*int(bf.Bitset.Length)
	sz = sz + 8*3
	return sz
}

func (bf *Filter) Upsert(item string) bool {
	if bf.Contains(item) {
		return true
	}

	bf.Add(item)
	bf.Cnt++
	return false
}

// Add adds an item to the Filter
func (bf *Filter) Add(item string) {
	for i := 0; i < bf.Hashes; i++ {
		index := bf.hash(item, i)
		bf.Bitset.Set(index)
	}
}

// Contains checks if an item might be in the Filter
func (bf *Filter) Contains(item string) bool {
	for i := 0; i < bf.Hashes; i++ {
		index := bf.hash(item, i)
		if !bf.Bitset.Query(index) {
			return false
		}
	}
	return true
}

// hash generates a hash for an item with a given seed
func (bf *Filter) hash(item string, seed int) uint {
	hasher := murmur3.New64WithSeed(uint32(seed))
	hasher.Write([]byte(item))
	return uint(hasher.Sum64() % uint64(bf.Size))
}
