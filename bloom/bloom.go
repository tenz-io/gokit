package bloom

import (
	"math"

	"github.com/spaolacci/murmur3"
)

type Filter interface {
	Add(data []byte)
	Exists(data []byte) bool
}

type filter struct {
	bitArray  []byte
	size      uint64
	numHashes int
}

// NewFilter creates a new bloom filter with the given number of elements and false positive probability.
// count is the number of elements to be inserted into the filter.
// p is the false positive probability.
func NewFilter(count uint64, p float64) Filter {
	arraySize, numHashes := optimalSize(count, p)
	return &filter{
		bitArray:  make([]byte, (arraySize+7)/8), // round up to the nearest byte
		size:      arraySize,
		numHashes: numHashes,
	}
}

func (f *filter) Add(data []byte) {
	for i := 0; i < f.numHashes; i++ {
		hashVal := f.hashWithSeed(data, uint32(i))
		index := hashVal % f.size
		f.setBit(index)
	}
}

func (f *filter) Exists(data []byte) bool {
	for i := 0; i < f.numHashes; i++ {
		hashVal := f.hashWithSeed(data, uint32(i))
		index := hashVal % f.size
		if !f.getBit(index) {
			return false
		}
	}
	return true
}

// hashWithSeed calculates the hash value of the data with the given seed.
func (f *filter) hashWithSeed(data []byte, seed uint32) uint64 {
	hasher := murmur3.New64WithSeed(seed)
	_, _ = hasher.Write(data)
	return hasher.Sum64()
}

// setBit sets the bit at the given index in the bit array.
func (f *filter) setBit(index uint64) {
	byteIndex := index / 8
	bitIndex := index % 8
	f.bitArray[byteIndex] |= 1 << bitIndex
}

// getBit returns the bit at the given index in the bit array.
func (f *filter) getBit(index uint64) bool {
	byteIndex := index / 8
	bitIndex := index % 8
	return f.bitArray[byteIndex]&(1<<bitIndex) != 0
}

// optimalSize calculates the optimal size of the bit array and the number of hash functions.
// n is the number of elements to be inserted into the filter.
// p is the false positive probability.
func optimalSize(n uint64, p float64) (arraySize uint64, hashNum int) {
	// calculate the optimal size of the bit array m
	m := -float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)
	// calculate the optimal number of hash functions k
	k := math.Ln2 * m / float64(n)
	return uint64(math.Ceil(m)), int(math.Ceil(k))
}
