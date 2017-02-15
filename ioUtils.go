package bunnystub

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
)

type IOWriter interface {
	io.Writer
}

var (
	// BufferSize defined the buffer size
	// about 2MB
	BufferSize = 0x6ffffff
	// Precision defined the precision about how many SECONDS will be waitted before
	// the reopen operation check the condition
	Precision = 1
	// WaitForClose wait for SECONDS then close the last file writer
	WartForClose = 10

	// ErrInternal defined the internal error
	ErrInternal = errors.New("error internal")
	// ErrInvalidArgument defined the invalid argument
	ErrInvalidArgument = errors.New("error argument invalid")
)

// NewIOWriter generate a iofilter writer with given ioManager
func NewIOWriter(m ioManager) (IOWriter, error) {
	if m == nil {
		return nil, ErrInvalidArgument
	}
	var writer = &fileWriter{
		event:     make(chan string, 2),
		buffer:    make(chan []byte, BufferSize),
		precision: time.Tick(time.Duration(Precision) * time.Second),
		manager:   m,
	}

	path, prefix, suffix := m.NameParts()
	name := path + prefix + time.Now().Format("200601021504") + suffix + ".log"
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	writer.file = file
	writer.started = make(chan int, 2)

	go writer.conditionWrite()

	<-writer.started
	<-writer.started
	return writer, nil
}

type fileWriter struct {
	file   *os.File
	event  chan string
	buffer chan []byte

	started   chan int
	precision <-chan time.Time
	size      int64
	version   int64
	manager   ioManager
}

func (w *fileWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.buffer <- p
	return n, nil
}

func (w *fileWriter) conditionWrite() {
	defer syscall.Sync()

	go func() {
		w.started <- 1
		for {
			if s, ok := w.manager.Enable(); ok {
				w.event <- s
			}
			<-w.precision
		}
	}()
	w.started <- 1

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

	file, err := os.OpenFile(path+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		if err == os.ErrNotExist {
			log.Println(err)
			w.file = os.Stderr
		}
	}
	w.file = file

	// Do the additional jobs like compresing the log file
	go func() {
		<-time.Tick(time.Second * time.Duration(WartForClose))
		oldFile.Seek(0, 0)
		info, _ := oldFile.Stat()
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

			cmpname := strings.TrimSuffix(info.Name(), ".log") + ".tar.gz"
			cmpfile, err := os.OpenFile(cmpname, os.O_RDWR|os.O_CREATE|os.O_APPEND, info.Mode())
			defer cmpfile.Close()
			if err != nil {
				log.Println(err)
				return
			}
			gw := gzip.NewWriter(cmpfile)
			defer gw.Close()
			tw := tar.NewWriter(gw)
			defer tw.Close()

			tw.WriteHeader(&tar.Header{
				Name: info.Name(),
				Mode: int64(info.Mode()),
				Size: info.Size(),
			})

			io.Copy(tw, oldFile)
		}
		oldFile.Close()
		os.Remove(info.Name())
	}()
}
