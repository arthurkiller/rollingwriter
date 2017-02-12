package rollingWriter

import (
	"archive/tar"
	"errors"
	"io"
	"log"
	"os"
	"time"
)

type IOFilterWriter interface {
	io.Writer
}

var (
	// BufferSize defined the buffer size
	// about 2MB
	BufferSize = 0x6ffffff
	// Precision defined the precision about how many SECONDS will be waitted before
	// the reopen operation check the condition
	Precision = 1

	// ErrInternal defined the internal error
	ErrInternal = errors.New("error internal")
	// ErrInvalidArgument defined the invalid argument
	ErrInvalidArgument = errors.New("error argument invalid")
)

// NewIOFilterWriter generate a iofilter writer with given ioManager
func NewIOFilterWriter(m ioManager) (IOFilterWriter, error) {
	if m == nil {
		return nil, ErrInvalidArgument
	}
	var writer = &fileWriter{
		event:     make(chan string, 4),
		buffer:    make(chan []byte, BufferSize),
		precision: time.Tick(time.Duration(Precision) * time.Second),
		manager:   m,
	}

	file, err := os.OpenFile(m.Path(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	writer.file = file

	go writer.conditionWrite()
	return writer, nil
}

// defer flush

type fileWriter struct {
	file   *os.File
	event  chan string
	buffer chan []byte

	precision <-chan time.Time
	size      int64
	version   int64
	manager   ioManager
}

func (w *fileWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	w.buffer <- p
	return
}

func (w *fileWriter) conditionWrite() {
	go func() {
		for {
			w.manager.Enable()
			<-w.precision
		}
	}()

	for {
		select {
		case path := <-w.event:
			w.reopen(path)
		case v := <-w.buffer:
			n, err := w.file.Write(v)
			// if ignore error then do nothing
			if err != nil && !w.manager.IgnoreOK() {
				// FIXME is this a good way?
				log.Println(err)
				w.file = os.Stderr
			}
			w.size += int64(n)
		}
	}
}

func (w *fileWriter) reopen(path string) {
	oldFile := w.file

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		if err == os.ErrNotExist {
			log.Println(err)
			w.file = os.Stderr
		}
	}
	w.file = file

	// Do the additional jobs like compresing the log file
	go func() {
		name := oldFile.Name()
		if w.manager.Compress() {
			// Do compress the log file
			// name the compressed file
			// delete the old file
			var source string
			_, err := oldFile.WriteString(source)
			if err != nil {
				log.Println(err)
				return
			}

			cmpname := name + ".tar.gz"
			cmpfile, err := os.OpenFile(cmpname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				log.Println(err)
				return
			}

			tw := tar.NewWriter(cmpfile)
			tw.WriteHeader(&tar.Header{
				Name: name,
				Mode: 0600,
				Size: int64(len(source)),
			})
			if _, err := tw.Write([]byte(source)); err != nil {
				log.Println(err)
				return
			}
			cmpfile.Close()
		}

		<-time.Tick(time.Minute * 5)
		oldFile.Close()
		//FIXME err is ignored!
	}()
}
