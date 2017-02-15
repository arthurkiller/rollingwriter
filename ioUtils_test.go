package bunnystub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newWriter() *fileWriter {
	f := &fileWriter{
		event:     make(chan string, 4),
		buffer:    make(chan []byte, 0x7fffff),
		precision: time.Tick(time.Duration(Precision) * time.Second),
		manager:   newioManager(),
	}
	return f
}

func TestNewIOWriter(t *testing.T) {
	m := newioManager()
	w, _ := NewIOWriter(m)
	n, err := w.Write([]byte("hello,world"))

	assert.Equal(t, 11, n)
	assert.Equal(t, nil, err)
}
