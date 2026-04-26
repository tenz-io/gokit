// Package bloom provides a space-efficient probabilistic data structure
// for testing set membership with a configurable false positive rate.
package bloom

import (
	"math"

	"github.com/spaolacci/murmur3"
)

// Filter is a Bloom filter for approximate set membership tests.
type Filter interface {
	Add(data []byte)
	AddString(s string)
	Exists(data []byte) bool
	ExistsString(s string) bool
	ApproxCount() uint64
	FalsePositiveRate() float64
}

type filter struct {
	bitArray  []byte
	size      uint64
	numHashes int
	count     uint64 // approximate number of elements inserted
}

// NewFilter creates a Bloom filter sized for n expected elements with false positive probability p.
// n must be > 0. p must be in (0, 1). Returns nil if parameters are invalid.
func NewFilter(n uint64, p float64) Filter {
	if n == 0 || p <= 0 || p >= 1 {
		return nil
	}
	m, k := optimalParams(n, p)
	return &filter{
		bitArray:  make([]byte, (m+7)/8),
		size:      m,
		numHashes: k,
	}
}

func (f *filter) Add(data []byte) {
	for i := 0; i < f.numHashes; i++ {
		f.setBit(f.hashWithSeed(data, uint32(i)) % f.size)
	}
	f.count++
}

func (f *filter) AddString(s string) { f.Add([]byte(s)) }

func (f *filter) Exists(data []byte) bool {
	for i := 0; i < f.numHashes; i++ {
		if !f.getBit(f.hashWithSeed(data, uint32(i)) % f.size) {
			return false
		}
	}
	return true
}

func (f *filter) ExistsString(s string) bool { return f.Exists([]byte(s)) }

// ApproxCount returns the approximate number of elements inserted.
func (f *filter) ApproxCount() uint64 { return f.count }

// FalsePositiveRate estimates the current false positive probability based on elements inserted.
func (f *filter) FalsePositiveRate() float64 {
	if f.count == 0 {
		return 0
	}
	return math.Pow(1 - math.Exp(-float64(f.numHashes)*float64(f.count)/float64(f.size)), float64(f.numHashes))
}

func (f *filter) hashWithSeed(data []byte, seed uint32) uint64 {
	h := murmur3.New64WithSeed(seed)
	_, _ = h.Write(data)
	return h.Sum64()
}

func (f *filter) setBit(idx uint64) {
	f.bitArray[idx/8] |= 1 << (idx % 8)
}

func (f *filter) getBit(idx uint64) bool {
	return f.bitArray[idx/8]&(1<<(idx%8)) != 0
}

func optimalParams(n uint64, p float64) (uint64, int) {
	m := -float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)
	k := math.Ln2 * m / float64(n)
	return uint64(math.Ceil(m)), int(math.Ceil(k))
}
