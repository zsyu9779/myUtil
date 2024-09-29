package hash

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkMurmurhash3(b *testing.B) {
	key := []byte("Hello, world!")
	seed := uint32(42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Murmurhash3(key, seed)
	}
}

func BenchmarkMurmurHash3_LongString(b *testing.B) {
	key := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
	seed := uint32(42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Murmurhash3(key, seed)
	}
}

func BenchmarkMurmurHash3_ShortStringDifferentSeed(b *testing.B) {
	key := []byte("Hello, world!")
	seed := uint32(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Murmurhash3(key, seed)
	}
}

func BenchmarkMurmurHash3_LongStringDifferentSeed(b *testing.B) {
	key := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
	seed := uint32(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Murmurhash3(key, seed)
	}
}

// Function to generate random strings
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// Function to test collision rate
func TestCollisionRate(t *testing.T) {
	const numSamples = 1000000
	const strLength = 10
	seed := uint32(42)

	hashMap := make(map[uint32]int)
	collisions := 0

	for i := 0; i < numSamples; i++ {
		sample := randomString(strLength)
		hash := Murmurhash3([]byte(sample), seed)
		if _, exists := hashMap[hash]; exists {
			collisions++
		}
		hashMap[hash] = 1
	}

	collisionRate := float64(collisions) / float64(numSamples) * 100.0
	fmt.Printf("Number of samples: %d\n", numSamples)
	fmt.Printf("Number of collisions: %d\n", collisions)
	fmt.Printf("Collision rate: %f%%\n", collisionRate)
}
