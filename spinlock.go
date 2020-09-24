package rollingwriter

import (
	"runtime"
	"sync/atomic"
)

// Locker is a spinlock implementation.
//
// A Locker must not be copied after first use.
type Locker struct {
	// The status of Locker.
	// `0` means unlocked, `1` means locked. `lock` will be `0`(zero value, unlocked) when new a Locker.
	lock uintptr
}

// Lock will try to get thel lock and yield the processor.
func (l *Locker) Lock() {
	for !atomic.CompareAndSwapUintptr(&l.lock, 0, 1) { // try to get the lock
		runtime.Gosched() // yield the processor here, allowing other goroutines to run
	}
}

// Unlock unlocks l.
func (l *Locker) Unlock() {
	atomic.StoreUintptr(&l.lock, 0)
}
