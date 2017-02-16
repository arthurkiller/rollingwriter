package bunnystub

import (
	"os"
	"time"
)

var dest *os.File

func newWriter() *fileWriter {
	m := newioManager()
	f := &fileWriter{
		event:     make(chan string, 4),
		buffer:    make(chan []byte, 0x7fffff),
		precision: time.Tick(time.Duration(Precision) * time.Second),
		manager:   m,
		started:   make(chan int, 2),
		file:      dest,
	}

	path, prefix, suffix := m.NameParts()
	name := path + prefix + time.Now().Format("200601021504") + suffix + ".log"
	dest, _ = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	go f.conditionWrite()
	<-f.started
	<-f.started
	return f
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
