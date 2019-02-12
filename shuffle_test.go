package eth2_shuffle

import (
	"golang.org/x/crypto/sha3"
	"testing"
)

func BenchmarkPermuteIndex(b *testing.B) {
	hash := sha3.New256()
	hashFn := func(in []byte) []byte {
		hash.Reset()
		hash.Write(in)
		return hash.Sum(nil)
	}

	// "random" seed for testing. Can be any 32 bytes.
	seed := [32]byte{123, 42}
	// list-size to test, test the 4M validator number
	listSize := uint64(4000000)
	// rounds of shuffling, constant in spec
	rounds := uint64(90)
	for i := uint64(0); i < uint64(b.N); i++ {
		PermuteIndex(hashFn, rounds, i % listSize, listSize, seed)
	}
}
