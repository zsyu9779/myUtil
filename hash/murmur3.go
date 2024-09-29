package hash

import "encoding/binary"

// Helper functions

// Rotate left (circular shift left)
func rotl32(x uint32, r int) uint32 {
	return (x << r) | (x >> (32 - r))
}

// Finalization mix
func fmix32(h uint32) uint32 {
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

func Murmurhash3(key []byte, seed uint32) uint32 {
	h1 := seed
	var c1 uint32 = 0xcc9e2d51
	var c2 uint32 = 0x1b873593

	// Body
	nblocks := len(key) / 4
	for i := 0; i < nblocks; i++ {
		k1 := binary.LittleEndian.Uint32(key[i*4:])

		k1 *= c1
		k1 = rotl32(k1, 15)
		k1 *= c2

		h1 ^= k1
		h1 = rotl32(h1, 13)
		h1 = h1*5 + 0xe6546b64
	}

	// Tail
	var k1 uint32
	tail := key[nblocks*4:]

	switch len(tail) {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = rotl32(k1, 15)
		k1 *= c2
		h1 ^= k1
	}

	// Finalization
	h1 ^= uint32(len(key))
	h1 = fmix32(h1)

	return h1
}
