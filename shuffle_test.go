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
	const listSize = 4000000
	// rounds of shuffling, constant in spec
	rounds := uint64(90)
	// benchmark!
	for i := uint64(0); i < uint64(b.N); i++ {
		PermuteIndex(hashFn, rounds, i % listSize, listSize, seed)
	}
}

func BenchmarkShuffleList(b *testing.B) {
	hash := sha3.New256()
	hashFn := func(in []byte) []byte {
		hash.Reset()
		hash.Write(in)
		return hash.Sum(nil)
	}

	// "random" seed for testing. Can be any 32 bytes.
	seed := [32]byte{123, 42}
	// list to test, test the 4M-sized validator list
	const listSize = 4000000
	testIndices := make([]uint64, listSize, listSize)
	// fill
	for i := uint64(0); i < listSize; i++ {
		testIndices[i] = i
	}
	// rounds of shuffling, constant in spec
	rounds := uint64(90)
	// benchmark!
	for i := uint64(0); i < uint64(b.N); i++ {
		ShuffleList(hashFn, testIndices, rounds, seed)
	}
}
