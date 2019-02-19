package eth2_shuffle

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func getStandardHashFn() HashFn {
	hash := sha256.New()
	hashFn := func(in []byte) []byte {
		hash.Reset()
		hash.Write(in)
		return hash.Sum(nil)
	}
	return hashFn
}

func readEncodedListInput(input string, requiredLen int64, lineIndex int) ([]uint64, error) {
	var itemStrs []string
	if input != "" {
		itemStrs = strings.Split(input, ":")
	} else {
		itemStrs = make([]string, 0)
	}
	if int64(len(itemStrs)) != requiredLen {
		return nil, fmt.Errorf("expected outputs length does not match list size on line %d\n", lineIndex)
	}
	items := make([]uint64, len(itemStrs), len(itemStrs))
	for i, itemStr := range itemStrs {
		item, err := strconv.ParseInt(itemStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("expected list item on line %d, item %d cannot be parsed\n", lineIndex, i)
		}
		items[i] = uint64(item)
	}
	return items, nil
}

func TestAgainstSpec(t *testing.T) {
	// Open CSV file
	f, err := os.Open("spec/tests.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		panic(err)
	}

	// constant in spec
	rounds := uint64(90)

	// Loop through lines & turn into object
	for lineIndex, line := range lines {

		parsedSeed, err := hex.DecodeString(line[0])
		if err != nil {
			t.Fatalf("seed on line %d cannot be parsed\n", lineIndex)
		}
		listSize, err := strconv.ParseInt(line[1], 10, 32)
		if err != nil {
			t.Fatalf("list size on line %d cannot be parsed\n", lineIndex)
		}
		inputItems, err := readEncodedListInput(line[2], listSize, lineIndex)
		expectedItems, err := readEncodedListInput(line[3], listSize, lineIndex)

		t.Run("", func(listSize uint64, shuffleIn []uint64, shuffleOut []uint64) func(st *testing.T) {
			return func(st *testing.T) {
				seed := [32]byte{}
				copy(seed[:], parsedSeed)
				// run every test case in parallel. Input data is copied, for loop won't mess it up.
				st.Parallel()

				hashFn := getStandardHashFn()

				st.Run("PermuteIndex", func (it *testing.T) {
					for i := uint64(0); i < listSize; i++ {
						// calculate the permuted index. (i.e. shuffle single index)
						permuted := PermuteIndex(hashFn, rounds, i, listSize, seed)
						// compare with expectation
						if shuffleIn[i] != shuffleOut[permuted] {
							it.FailNow()
						}
					}
				})

				st.Run("UnpermuteIndex", func (it *testing.T) {
					// for each index, test un-permuting
					for i := uint64(0); i < listSize; i++ {
						// calculate the un-permuted index. (i.e. un-shuffle single index)
						unpermuted := UnpermuteIndex(hashFn, rounds, i, listSize, seed)
						// compare with expectation
						if shuffleOut[i] != shuffleIn[unpermuted] {
							it.FailNow()
						}
					}
				})

				st.Run("ShuffleList", func (it *testing.T) {
					// create input, this slice will be shuffled.
					testInput := make([]uint64, listSize, listSize)
					copy(testInput, shuffleIn)
					// shuffle!
					ShuffleList(hashFn, testInput, rounds, seed)
					// compare shuffled list to expected output
					for i := uint64(0); i < listSize; i++ {
						if testInput[i] != shuffleOut[i] {
							it.FailNow()
						}
					}
				})

				st.Run("UnshuffleList", func (it *testing.T) {
					// create input, this slice will be un-shuffled.
					testInput := make([]uint64, listSize, listSize)
					copy(testInput, shuffleOut)
					// un-shuffle!
					UnshuffleList(hashFn, testInput, rounds, seed)
					// compare shuffled list to original input
					for i := uint64(0); i < listSize; i++ {
						if testInput[i] != shuffleIn[i] {
							it.FailNow()
						}
					}
				})
			}
		}(uint64(listSize), inputItems, expectedItems))
	}
}

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

