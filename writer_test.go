package rollingwriter

import (
	"io"
	"math/rand"
	"os"
	"sync"
	"testing"
)

func clean() {
	os.Remove("./test/unittest*")
	os.Remove("./test")
}

func newWriter() *Writer {
	cfg := NewDefaultConfig()
	cfg.LogPath = "./test"
	cfg.FileName = "unittest"
	cfg.Asynchronous = false
	cfg.Lock = false
	w, _ := NewWriterFromConfig(&cfg)
	return w.(*Writer)
}

func newLockedWriter() *LockedWriter {
	cfg := NewDefaultConfig()
	cfg.LogPath = "./test"
	cfg.FileName = "unittest"
	cfg.Asynchronous = false
	cfg.Lock = true
	w, _ := NewWriterFromConfig(&cfg)
	return w.(*LockedWriter)
}

func newAsynWriter() *AsynchronousWriter {
	cfg := NewDefaultConfig()
	cfg.LogPath = "./test"
	cfg.FileName = "unittest"
	cfg.Asynchronous = true
	w, _ := NewWriterFromConfig(&cfg)
	return w.(*AsynchronousWriter)
}

// TODO
func newBufferWriter() *BufferWriter {
	return &BufferWriter{}
}

func TestWrite(t *testing.T) {
	var writer io.WriteCloser
	var c int = 10000
	var l int = 1024
	wg := sync.WaitGroup{}

	writer = newWriter()
	for i := 0; i < c; i++ {
		wg.Add(1)
		bf := make([]byte, l)
		rand.Read(bf)
		go func() {
			writer.Write(bf)
			defer wg.Done()
		}()
	}
	wg.Wait()
	writer.Close()
	clean()

	writer = newLockedWriter()
	for i := 0; i < c; i++ {
		wg.Add(1)
		bf := make([]byte, l)
		rand.Read(bf)
		go func() {
			writer.Write(bf)
			defer wg.Done()
		}()
	}
	wg.Wait()
	writer.Close()
	clean()

	writer = newAsynWriter()
	for i := 0; i < c; i++ {
		wg.Add(1)
		bf := make([]byte, l)
		rand.Read(bf)
		go func() {
			writer.Write(bf)
			defer wg.Done()
		}()
	}
	wg.Wait()
	writer.Close()
	clean()

	// TODO
	//writer = newBufferWriter()
	//for i := 0; i < c; i++ {
	//	wg.Add(1)
	//	bf := make([]byte, l)
	//	rand.Read(bf)
	//	go func() {
	//		writer.Write(bf)
	//		defer wg.Done()
	//	}()
	//}
	//wg.Wait()
}

func TestReopen(t *testing.T) {
}
