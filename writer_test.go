package bunnystub

import (
	"os"
	"sync"
	"testing"
)

var dest *os.File

func newTWriter() IOWriter {
	wr, err := NewIOWriter("/dev/null", TimeRotate, WithConcurrency(100, true))
	if err != nil {
		panic(err)
	}
	return wr
}

func newVWriter() IOWriter {
	wr, err := NewIOWriter("/dev/null", VolumeRotate, WithVolumeSize("1024b"))
	if err != nil {
		panic(err)
	}
	return wr
	//m := newManager()

	//f := &fileWriter{
	//	event:     make(chan string, 4),
	//	buffer:    make(chan []byte, 0x7fffff),
	//	precision: time.Tick(time.Duration(Precision) * time.Second),
	//	manager:   m,
	//	started:   make(chan int, 2),
	//	file:      dest,
	//}

	//path, prefix, suffix := m.NameParts()
	//name := path + prefix + time.Now().Format("200601021504") + suffix + ".log"
	//dest, _ = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	//go f.conditionWrite()
	//<-f.started
	//<-f.started
	//return f
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

//func TestNewIOWriter(t *testing.T) {
//	m := newioManager()
//	w, _ := NewIOWriter(m)
//	n, err := w.Write([]byte("hello,world"))
//
//	assert.Equal(t, 11, n)
//	assert.Equal(t, nil, err)
//}
//
//func TestWrite(t *testing.T) {
//	dest, _ = os.Create("test")
//	//defer func() {
//	//	dest.Close()
//	//	os.Remove("test")
//	//}()
//
//	w := newWriter()
//	n, err := w.Write([]byte("hello,world"))
//	assert.Equal(t, nil, err)
//	assert.Equal(t, 11, n)
//
//	buf := make([]byte, 11)
//	f, err := os.Open("test")
//	assert.Equal(t, nil, err)
//	n, err = f.Read(buf)
//	assert.Equal(t, nil, err)
//	assert.Equal(t, 11, n)
//
//	assert.Equal(t, string(buf), "hello,world")
//}
