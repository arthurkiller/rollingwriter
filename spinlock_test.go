package rollingwriter

import (
	"sync"
	"testing"
)

func TestSpinLock(t *testing.T) {
	var counter int
	var wg sync.WaitGroup
	spinLock := &Locker{}

	startCh := make(chan struct{})

	wg.Add(5000)
	for i := 0; i < 5000; i++ {
		go func() {
			defer wg.Done()

			// waitting start
			<-startCh

			for j := 0; j < 10000; j++ {
				spinLock.Lock()
				counter++
				spinLock.Unlock()
			}
		}()
	}

	close(startCh)
	wg.Wait()

	if counter != 50000000 {
		t.Errorf("Counter should be 50000000, but got %d", counter)
	}
}
