package bunnystub

import (
	"compress/gzip"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"sync"
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
	WartForClose = 5

	// ErrInternal defined the internal error
	ErrInternal = errors.New("error internal")
	// ErrInvalidArgument defined the invalid argument
	ErrInvalidArgument = errors.New("error argument invalid")
)

// NewIOWriter generate a iofilter writer with given ioManager
func NewIOWriter(ops ...Option) (IOWriter, error) {
	m := newIOManager(ops...)

	if m.LockFree() {
		var writer = &lockFreeWriter{
			event:     make(chan string, 2),
			buffer:    make(chan []byte, BufferSize),
			precision: time.Tick(time.Duration(Precision) * time.Second),
			manager:   m,
		}

		path, prefix, suffix := m.NameParts()
		name := path + prefix + suffix
		file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, m.FileMode())
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

	var writer = &fileWriter{event: make(chan string, 2), precision: time.Tick(time.Duration(Precision) * time.Second), manager: m}

	path, prefix, suffix := m.NameParts()
	name := path + prefix + suffix
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, m.FileMode())
	if err != nil {
		return nil, err
	}
	writer.file = file
	writer.started = make(chan int, 2)

	go writer.conditionManager()

	// wait for conditionwrite start
	<-writer.started
	<-writer.started

	return writer, nil
}

type fileWriter struct {
	file      *os.File
	l         sync.RWMutex
	manager   ioManager
	started   chan int
	precision <-chan time.Time
	event     chan string
	close     chan byte
}

func (w *fileWriter) Write(s []byte) (int, error) {
	w.l.RLock()
	defer w.l.RUnlock()
	return w.file.Write(s)
}

func (w *fileWriter) Close() error {
	close(w.close)

	return w.file.Close()
}

func (w *fileWriter) conditionManager() {
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
		case lastname := <-w.event:
			oldFile := w.file
			oldFileName := oldFile.Name()

			err := os.Rename(oldFileName, lastname)
			if err != nil {
				log.Println("error in rename file", err)
				return
			}

			// open & swap the file
			file, err := os.OpenFile(oldFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, w.manager.FileMode())
			if err != nil {
				log.Println("error in reopen file", err)
				return
			}
			w.file.Sync()
			w.file = file

			// Do additional jobs like compresing the log file
			go func() {
				w.l.Lock()
				defer w.l.Unlock()
				if w.manager.Compress() {
					// Do compress the log file
					// name the compressed file
					// delete the old file
					cmpname := strings.TrimSuffix(lastname, ".log") + ".gz"
					cmpfile, err := os.OpenFile(cmpname, os.O_RDWR|os.O_CREATE|os.O_APPEND, w.manager.FileMode())
					defer cmpfile.Close()
					if err != nil {
						log.Println("error in reopen additional goroution", err)
						return
					}
					gw := gzip.NewWriter(cmpfile)
					defer gw.Close()

					oldFile.Seek(0, 0)
					io.Copy(gw, oldFile)
					os.Remove(lastname) //remove *.log
				}
				oldFile.Close()
			}()
		case <-w.close:
			break
		}
	}
}

type lockFreeWriter struct {
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

func (w *lockFreeWriter) Write(s []byte) (int, error) {
	n := len(s)
	p := make([]byte, n)
	copy(p, s)
	w.buffer <- p
	return n, nil
}

func (w *lockFreeWriter) Close() error {
	close(w.close)

	return w.file.Close()
}

func (w *lockFreeWriter) conditionWrite() {
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

func (w *lockFreeWriter) reopen(lastname string) {
	oldFile := w.file
	oldFileName := oldFile.Name()
	err := os.Rename(oldFileName, lastname)
	if err != nil {
		log.Println("error in rename file", err)
		return
	}

	// open & swap the file
	file, err := os.OpenFile(oldFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, w.manager.FileMode())
	if err != nil {
		log.Println("error in reopen file", err)
		return
	}
	w.file.Sync()
	w.file = file

	// Do additional jobs like compresing the log file
	go func() {
		<-time.Tick(time.Second * time.Duration(WartForClose))
		if w.manager.Compress() {
			// Do compress the log file
			// name the compressed file
			// delete the old file
			cmpname := strings.TrimSuffix(lastname, ".log") + ".gz"
			cmpfile, err := os.OpenFile(cmpname, os.O_RDWR|os.O_CREATE|os.O_APPEND, w.manager.FileMode())
			defer cmpfile.Close()
			if err != nil {
				log.Println("error in reopen additional goroution", err)
				return
			}
			gw := gzip.NewWriter(cmpfile)
			defer gw.Close()

			oldFile.Seek(0, 0)
			io.Copy(gw, oldFile)
			os.Remove(lastname)
		}
		oldFile.Close()
	}()
}
