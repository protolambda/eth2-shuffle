package eth2_shuffle

import "encoding/binary"

type Interface interface {
	// Len is the number of elements in the collection.
	Len() int

	// Swap swaps the elements with indexes i and j.
	Swap(i, j int)
}

type Uint64Slice []uint64

func (x Uint64Slice) Len() int      { return len(x) }
func (x Uint64Slice) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

type Uint32Slice []uint32

func (x Uint32Slice) Len() int      { return len(x) }
func (x Uint32Slice) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

// Shuffles the list
func ShuffleListSwap(hashFn HashFn, data Interface, rounds uint8, seed [32]byte) {
	innerShuffleListSwap(hashFn, data, rounds, seed, true)
}

// Un-shuffles the list
func UnshuffleListSwap(hashFn HashFn, data Interface, rounds uint8, seed [32]byte) {
	innerShuffleListSwap(hashFn, data, rounds, seed, false)
}

// Shuffles or unshuffles, depending on the `dir` (true for shuffling, false for unshuffling
func innerShuffleListSwap(hashFn HashFn, data Interface, rounds uint8, seed [32]byte, dir bool) {
	if rounds == 0 {
		return
	}
	listSize := uint64(data.Len())
	if listSize == 0 {
		return
	}
	buf := make([]byte, hTotalSize, hTotalSize)
	r := uint8(0)
	if !dir {
		// Start at last round.
		// Iterating through the rounds in reverse, un-swaps everything, effectively un-shuffling the list.
		r = rounds - 1
	}
	// Seed is always the first 32 bytes of the hash input, we never have to change this part of the buffer.
	copy(buf[:hSeedSize], seed[:])
	for {
		// spec: pivot = bytes_to_int(hash(seed + int_to_bytes1(round))[0:8]) % list_size
		// This is the "int_to_bytes1(round)", appended to the seed.
		buf[hSeedSize] = r
		// Seed is already in place, now just hash the correct part of the buffer, and take a uint64 from it,
		//  and modulo it to get a pivot within range.
		pivot := binary.LittleEndian.Uint64(hashFn(buf[:hPivotViewSize])[:8]) % listSize

		// Split up the for-loop in two:
		//  1. Handle the part from 0 (incl) to pivot (incl). This is mirrored around (pivot / 2)
		//  2. Handle the part from pivot (excl) to N (excl). This is mirrored around ((pivot / 2) + (size/2))
		// The pivot defines a split in the array, with each of the splits mirroring their data within the split.
		// Print out some example even/odd sized index lists, with some even/odd pivots,
		//  and you can deduce how the mirroring works exactly.
		// Note that the mirror is strict enough to not consider swapping the index @mirror with itself.
		mirror := (pivot + 1) >> 1
		// Since we are iterating through the "positions" in order, we can just repeat the hash every 256th position.
		// No need to pre-compute every possible hash for efficiency like in the example code.
		// We only need it consecutively (we are going through each in reverse order however, but same thing)
		//
		// spec: source = hash(seed + int_to_bytes1(round) + int_to_bytes4(position // 256))
		// - seed is still in 0:32 (excl., 32 bytes)
		// - round number is still in 32
		// - mix in the position for randomness, except the last byte of it,
		//     which will be used later to select a bit from the resulting hash.
		// We start from the pivot position, and work back to the mirror position (of the part left to the pivot).
		// This makes us process each pear exactly once (instead of unnecessarily twice, like in the spec)
		binary.LittleEndian.PutUint32(buf[hPivotViewSize:], uint32(pivot>>8))
		source := hashFn(buf)
		byteV := source[(pivot&0xff)>>3]
		for i, j := int(0), pivot; uint64(i) < mirror; i, j = i+1, j-1 {
			// The pair is i,j. With j being the bigger of the two, hence the "position" identifier of the pair.
			// Every 256th bit (aligned to j).
			if j&0xff == 0xff {
				// just overwrite the last part of the buffer, reuse the start (seed, round)
				binary.LittleEndian.PutUint32(buf[hPivotViewSize:], uint32(j>>8))
				source = hashFn(buf)
			}
			// Same trick with byte retrieval. Only every 8th.
			if j&0x7 == 0x7 {
				byteV = source[(j&0xff)>>3]
			}
			bitV := (byteV >> (j & 0x7)) & 0x1

			if bitV == 1 {
				// swap the pair items
				data.Swap(i, int(j))
			}
		}
		// Now repeat, but for the part after the pivot.
		mirror = (pivot + listSize + 1) >> 1
		end := listSize - 1
		// Again, seed and round input is in place, just update the position.
		// We start at the end, and work back to the mirror point.
		// This makes us process each pear exactly once (instead of unnecessarily twice, like in the spec)
		binary.LittleEndian.PutUint32(buf[hPivotViewSize:], uint32(end>>8))
		source = hashFn(buf)
		byteV = source[(end&0xff)>>3]
		for i, j := int(pivot+1), end; uint64(i) < mirror; i, j = i+1, j-1 {
			// Exact same thing (copy of above loop body)
			//--------------------------------------------
			// The pair is i,j. With j being the bigger of the two, hence the "position" identifier of the pair.
			// Every 256th bit (aligned to j).
			if j&0xff == 0xff {
				// just overwrite the last part of the buffer, reuse the start (seed, round)
				binary.LittleEndian.PutUint32(buf[hPivotViewSize:], uint32(j>>8))
				source = hashFn(buf)
			}
			// Same trick with byte retrieval. Only every 8th.
			if j&0x7 == 0x7 {
				byteV = source[(j&0xff)>>3]
			}
			bitV := (byteV >> (j & 0x7)) & 0x1

			if bitV == 1 {
				// swap the pair items
				data.Swap(i, int(j))
			}
			//--------------------------------------------
		}
		// go forwards?
		if dir {
			// -> shuffle
			r++
			if r == rounds {
				break
			}
		} else {
			if r == 0 {
				break
			}
			// -> un-shuffle
			r--
		}
	}
}
