package bunnystub

import (
	"os"
	"sync"
	"testing"
)

var dest *os.File

func newWriter() Writer {
}

func newLockedWriter() LockedWriter {
	return wr
}

func newAsynWriter() AsynchronousWriter {
	return wr
}

func newBufferWriter() BufferWriter {
	return wr
}

func TestWrite(t *testing.T) {
	writer := newTWriter()
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			writer.Write([]byte("hi"))
			defer wg.Done()
		}()
	}
	wg.Wait()

	writer = newVWriter()
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			writer.Write([]byte("hi"))
			defer wg.Done()
		}()
	}
	wg.Wait()
}
