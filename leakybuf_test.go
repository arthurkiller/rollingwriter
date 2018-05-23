package bunnystub

import "testing"

func TestLeakybuf(t *testing.T) {
	b := NewLeakyBuf(100, 100)
	go func() {
		bf := b.Get()
		b.Put(bf)
	}()
}
