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
	Close() error
}

var (
	// BufferSize defined the buffer size
	// about 2MB
	BufferSize = 0x6ffffff
	// Precision defined the precision about how many SECONDS will be waitted before
	// the reopen operation check the condition
	Precision = 1
	// WaitForClose wait for SECONDS then close the last file writer
	WartForClose = 1

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
	name := path + prefix + suffix + ".log"
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	writer.file = file
	writer.started = make(chan int, 2)

	go writer.conditionWrite()

	// wait for conditionwrite start
	<-writer.started
	<-writer.started

	return writer, nil
}

type fileWriter struct {
	file   *os.File
	event  chan string
	buffer chan []byte
	close  chan byte

	started   chan int
	precision <-chan time.Time
	size      int64
	version   int64
	manager   ioManager
}

func (w *fileWriter) Write(s []byte) (int, error) {
	n := len(s)
	p := make([]byte, n)
	copy(p, s)
	w.buffer <- p
	return n, nil
}

func (w *fileWriter) Close() error {
	w.file.Close()
	close(w.close)

	return nil
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
				i, _ := w.file.Stat()
				log.Println("err in condition write", err, i)
				w.file = os.Stderr
			}
			w.size += int64(n)
		case <-w.close:
			break
		}
	}
}

func (w *fileWriter) reopen(lastname string) {
	oldFile := w.file
	oldFileName := oldFile.Name()
	i, _ := oldFile.Stat()
	oldSize := i.Size()
	err := os.Rename(oldFileName, "./"+lastname)
	if err != nil {
		log.Println("error in rename file", err)
		return
	}

	// open & swap the file
	file, err := os.OpenFile(oldFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println("error in reopen file", err)
		return
	}
	w.file = file

	// Do additional jobs like compresing the log file
	go func() {
		<-time.Tick(time.Second * time.Duration(WartForClose))
		if w.manager.Compress() {
			// Do compress the log file
			// name the compressed file
			// delete the old file
			cmpname := strings.TrimSuffix(lastname, ".log") + ".tar.gz"
			cmpfile, err := os.OpenFile(cmpname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			defer cmpfile.Close()
			if err != nil {
				log.Println("error in reopen additional goroution", err)
				return
			}
			gw := gzip.NewWriter(cmpfile)
			defer gw.Close()
			tw := tar.NewWriter(gw)
			defer tw.Close()

			tw.WriteHeader(&tar.Header{
				Name: oldFileName,
				Mode: 0644,
				Size: oldSize,
			})

			oldFile.Seek(0, 0)
			io.Copy(tw, oldFile)
			os.Remove(lastname)
		}
		oldFile.Close()
	}()
}
