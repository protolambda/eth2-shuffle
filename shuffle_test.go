package eth2_shuffle

import (
	"golang.org/x/crypto/sha3"
	"testing"
)

func getStandardHashFn() HashFn {
	hash := sha3.New256()
	hashFn := func(in []byte) []byte {
		hash.Reset()
		hash.Write(in)
		return hash.Sum(nil)
	}
	return hashFn
}

func TestPermuteIndex(t *testing.T) {
	hashFn := getStandardHashFn()
	// "random" seed for testing. Can be any 32 bytes.
	seed := [32]byte{123, 42}
	// rounds of shuffling, constant in spec
	rounds := uint64(90)
	//
	listSize := uint64(4000000)
	// TODO parametrize
	i := uint64(10)
	permuted := PermuteIndex(hashFn, rounds, i % listSize, listSize, seed)
	// TODO check each parametrized number for correct shuffled output?
	if permuted != 123 {
		t.Fail()
	}
}

func TestUnpermuteIndex(t *testing.T) {
	// TODO
}

func TestDoUndoPermuteIndex(t *testing.T) {
	// TODO
}

func TestShuffleList(t *testing.T) {
	// TODO
}

func TestUnshuffleList(t *testing.T) {
	// TODO
}

func TestDoUndoShuffleList(t *testing.T) {
	// TODO
}

// TODO also test with test vectors from ETH 2.0 tests repo.

func BenchPermuteIndex(listSize uint64, b *testing.B) {
	hashFn := getStandardHashFn()
	// "random" seed for testing. Can be any 32 bytes.
	seed := [32]byte{123, 42}
	// rounds of shuffling, constant in spec
	rounds := uint64(90)
	// benchmark!
	for i := uint64(0); i < uint64(b.N); i++ {
		PermuteIndex(hashFn, rounds, i % listSize, listSize, seed)
	}
}

func BenchIndexComparison(listSize uint64, b *testing.B) {
	hashFn := getStandardHashFn()
	// "random" seed for testing. Can be any 32 bytes.
	seed := [32]byte{123, 42}
	// rounds of shuffling, constant in spec
	rounds := uint64(90)
	// benchmark!
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < listSize; j++ {
			PermuteIndex(hashFn, rounds, j, listSize, seed)
		}
	}
}

func BenchShuffleList(listSize uint64, b *testing.B) {
	hashFn := getStandardHashFn()
	// "random" seed for testing. Can be any 32 bytes.
	seed := [32]byte{123, 42}
	// list to test
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

func BenchmarkPermuteIndex4M(b *testing.B) {BenchPermuteIndex(4000000, b)}
func BenchmarkPermuteIndex40K(b *testing.B) {BenchPermuteIndex(40000, b)}
func BenchmarkPermuteIndex400(b *testing.B) {BenchPermuteIndex(400, b)}
//func BenchmarkPermuteIndexComparison4M(b *testing.B) {BenchIndexComparison(4000000, b)}
func BenchmarkPermuteIndexComparison40K(b *testing.B) {BenchIndexComparison(40000, b)}
func BenchmarkPermuteIndexComparison400(b *testing.B) {BenchIndexComparison(400, b)}
func BenchmarkShuffleList4M(b *testing.B) {BenchShuffleList(4000000, b)}
func BenchmarkShuffleList40K(b *testing.B) {BenchShuffleList(40000, b)}
func BenchmarkShuffleList400(b *testing.B) {BenchShuffleList(400, b)}

