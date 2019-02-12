package eth2_shuffle

import "encoding/binary"

type HashFn func(input []byte) []byte

/*
Return `p(index)` in a pseudorandom permutation `p` of `0...list_size-1` with ``seed`` as entropy.

    Utilizes 'swap or not' shuffling found in
    https://link.springer.com/content/pdf/10.1007%2F978-3-642-32009-5_1.pdf
    See the 'generalized domain' algorithm on page 3.

Eth 2.0 spec implementation here:
	https://github.com/ethereum/eth2.0-specs/blob/dev/specs/core/0_beacon-chain.md#get_permuted_index
 */
func PermuteIndex(hashFn HashFn, rounds uint64, index uint64, listSize uint64, seed [32]byte) uint64 {
	buf := make([]byte, 32 + 5 + 1, 32 + 5 + 1)
	for i := uint64(0); i < rounds; i++ {
		// spec: pivot = bytes_to_int(hash(seed + int_to_bytes1(round))[0:8]) % list_size
		copy(buf[:32], seed[:])
		buf[32] = byte(i)
		pivot := binary.LittleEndian.Uint64(hashFn(buf[:32 + 1])[:8]) % listSize
		// spec: flip = (pivot - index) % list_size
		// add extra list_size to prevent underflows.
		flip := (pivot + (listSize - index)) % listSize
		// spec: position = max(index, flip)
		position := index
		if flip > position {
			position = flip
		}
		// spec: source = hash(seed + int_to_bytes1(round) + int_to_bytes4(position // 256))
		// - seed is still in 0:32
		// - i is still in 33
		binary.LittleEndian.PutUint32(buf[32 + 1:32 + 1 + 5], uint32(position >> 8))
		source := hashFn(buf)
		// spec: byte = source[(position % 256) // 8]
		byteV := source[(position & 0xff) >> 3]
		// spec: bit = (byte >> (position % 8)) % 2
		bitV := (byteV >> (position & 0x7)) & 0x1
		// if bitV, flip.
		if bitV == 1 {
			index = flip
		}
	}
	return index
}

