package g_es

import (
	"math/rand"
	"testing"
)

func TestLo(t *testing.T) {
	var front []int
	var back []int
	for i := 0; i < 5; i++ {
		num := 1 + rand.Intn(36)
		front = append(front, num)
	}
	for i := 0; i < 2; i++ {
		num := 1 + rand.Intn(12)
		back = append(back, num)
	}
	t.Logf("front: %v, back: %v", front, back)
}
