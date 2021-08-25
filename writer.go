package rollingwriter

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Writer provide a synchronous file writer
// if Lock is set true, write will be guaranteed by lock
type Writer struct {
	m             Manager
	file          *os.File
	absPath       string
	fire          chan string
	cf            *Config
	rollingfilech chan string
}

// LockedWriter provide a synchronous writer with lock
// write operate will be guaranteed by lock
type LockedWriter struct {
	Writer
	sync.Mutex
}

// AsynchronousWriter provide a asynchronous writer with the writer to confirm the write
type AsynchronousWriter struct {
	Writer
	ctx     chan int
	queue   chan []byte
	errChan chan error
	closed  int32
	wg      sync.WaitGroup
}

// BufferWriter merge some write operations into one.
type BufferWriter struct {
	Writer
	buf     *[]byte // store the pointer for atomic opertaion
	swaping int32

	lockBuf Locker // protect the buffer by spinlock
}

// buffer pool for asynchronous writer
var _asyncBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, BufferSize)
	},
}

// NewWriterFromConfig generate the rollingWriter with given config
func NewWriterFromConfig(c *Config) (RollingWriter, error) {
	// makeup log path and create
	if c.LogPath == "" || c.FileName == "" {
		return nil, ErrInvalidArgument
	}

	// make dir for path if not exist
	if err := os.MkdirAll(c.LogPath, 0700); err != nil {
		return nil, err
	}

	filepath := LogFilePath(c)
	// open the file and get the FD
	file, err := os.OpenFile(filepath, DefaultFileFlag, DefaultFileMode)
	if err != nil {
		return nil, err
	}

	// Start the Manager
	mng, err := NewManager(c)
	if err != nil {
		return nil, err
	}

	var rollingWriter RollingWriter
	writer := Writer{
		m:       mng,
		file:    file,
		absPath: filepath,
		fire:    mng.Fire(),
		cf:      c,
	}

	if c.MaxRemain > 0 {
		writer.rollingfilech = make(chan string, c.MaxRemain)
		dir, err := ioutil.ReadDir(c.LogPath)
		if err != nil {
			return nil, err
		}

		files := make([]string, 0, 10)
		for _, fi := range dir {
			if fi.IsDir() {
				continue
			}

			fileName := c.FileName + ".log."
			if strings.Contains(fi.Name(), fileName) {
				fileSuffix := path.Ext(fi.Name())
				if len(fileSuffix) > 1 {
					_, err := time.Parse(c.TimeTagFormat, fileSuffix[1:])
					if err == nil {
						files = append(files, fi.Name())
					}
				}
			}
		}
		sort.Slice(files, func(i, j int) bool {
			fileSuffix1 := path.Ext(files[i])
			fileSuffix2 := path.Ext(files[j])
			t1, _ := time.Parse(c.TimeTagFormat, fileSuffix1[1:])
			t2, _ := time.Parse(c.TimeTagFormat, fileSuffix2[1:])
			return t1.Before(t2)
		})

		for _, file := range files {
		retry:
			select {
			case writer.rollingfilech <- path.Join(c.LogPath, file):
			default:
				writer.DoRemove()
				goto retry // remove the file and retry
			}
		}
	}

	switch c.WriterMode {
	case "none":
		rollingWriter = &writer
	case "lock":
		rollingWriter = &LockedWriter{
			Writer: writer,
		}
	case "async":
		wr := &AsynchronousWriter{
			ctx:     make(chan int),
			queue:   make(chan []byte, QueueSize),
			errChan: make(chan error),
			wg:      sync.WaitGroup{},
			closed:  0,
			Writer:  writer,
		}
		// start the asynchronous writer
		wr.wg.Add(1)
		go wr.writer()
		wr.wg.Wait()
		rollingWriter = wr
	case "buffer":
		// bufferWriterThershould unit is Byte
		bf := make([]byte, 0, c.BufferWriterThershould*2)
		rollingWriter = &BufferWriter{
			Writer:  writer,
			buf:     &bf,
			swaping: 0,
		}
	default:
		return nil, ErrInvalidArgument
	}
	return rollingWriter, nil
}

// NewWriter generate the rollingWriter with given option
func NewWriter(ops ...Option) (RollingWriter, error) {
	cfg := NewDefaultConfig()
	for _, opt := range ops {
		opt(&cfg)
	}
	return NewWriterFromConfig(&cfg)
}

// NewWriterFromConfigFile generate the rollingWriter with given config file
func NewWriterFromConfigFile(path string) (RollingWriter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := NewDefaultConfig()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(buf, &cfg); err != nil {
		return nil, err
	}
	return NewWriterFromConfig(&cfg)
}

// DoRemove will delete the oldest file
func (w *Writer) DoRemove() {
	select {
	case file := <-w.rollingfilech:
		// remove the oldest file
		if err := os.Remove(file); err != nil {
			log.Println("error in remove log file", file, err)
		}
	}
}

// CompressFile compress log file write into .gz and remove source file
func (w *Writer) CompressFile(oldfile *os.File, cmpname string) error {
	cmpfile, err := os.OpenFile(cmpname, DefaultFileFlag, DefaultFileMode)
	defer cmpfile.Close()
	if err != nil {
		return err
	}
	gw := gzip.NewWriter(cmpfile)
	defer gw.Close()

	if _, err = oldfile.Seek(0, 0); err != nil {
		return err
	}

	if _, err = io.Copy(gw, oldfile); err != nil {
		if errR := os.Remove(cmpname); errR != nil {
			return errR
		}
		return err
	}
	return os.Remove(cmpname + ".tmp") // remove *.log.tmp file
}

// AsynchronousWriterErrorChan return the error channel for asyn writer
func AsynchronousWriterErrorChan(wr RollingWriter) (chan error, error) {
	if w, ok := wr.(*AsynchronousWriter); ok {
		return w.errChan, nil
	}
	return nil, ErrInvalidArgument
}

// Reopen do the rotate, open new file and swap FD then trate the old FD
func (w *Writer) Reopen(file string) error {
	if w.cf.FilterEmptyBackup {
		fileInfo, err := w.file.Stat()
		if err != nil {
			return err
		}

		if fileInfo.Size() == 0 {
			return nil
		}
	}

	if err := os.Rename(w.absPath, file); err != nil {
		return err
	}
	newfile, err := os.OpenFile(w.absPath, DefaultFileFlag, DefaultFileMode)
	if err != nil {
		return err
	}

	// swap the unsafe pointer
	oldfile := atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&w.file)), unsafe.Pointer(newfile))

	go func() {
		defer (*os.File)(oldfile).Close()
		if w.cf.Compress {
			if err := os.Rename(file, file+".tmp"); err != nil {
				log.Println("error in compress rename tempfile", err)
				return
			}
			if err := w.CompressFile((*os.File)(oldfile), file); err != nil {
				log.Println("error in compress log file", err)
				return
			}
		}

		if w.cf.MaxRemain > 0 {
		retry:
			select {
			case w.rollingfilech <- file:
			default:
				w.DoRemove()
				goto retry // remove the file and retry
			}
		}
	}()
	return nil
}

func (w *Writer) Write(b []byte) (int, error) {
	var ok = false
	for ! ok {
		select {
		case filename := <-w.fire:
			if err := w.Reopen(filename); err != nil {
				return 0, err
			}
		default:
			ok = true
		}
	}

	fp := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&w.file)))
	file := (*os.File)(fp)
	return file.Write(b)
}

func (w *LockedWriter) Write(b []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()

	var ok = false
	for ! ok {
		select {
		case filename := <-w.fire:
			if err := w.Reopen(filename); err != nil {
				return 0, err
			}
		default:
			ok = true
		}
	}

	n, err = w.file.Write(b)
	return
}

// Only when the error channel is empty, otherwise nothing will write and the last error will be return
// return the error channel
func (w *AsynchronousWriter) Write(b []byte) (int, error) {
	if atomic.LoadInt32(&w.closed) == 0 {
		var ok = false
		for ! ok {
			select {
			case filename := <-w.fire:
				if err := w.Reopen(filename); err != nil {
					return 0, err
				}
			case err := <-w.errChan:
				// NOTE this error caused by last write maybe ignored
				return 0, err

			default:
				ok = true
			}
		}

		w.queue <- append(_asyncBufferPool.Get().([]byte)[0:0], b...)[:len(b)]
		return len(b), nil
	}
	return 0, ErrClosed
}

// writer do the asynchronous write independently
// Take care of reopen, I am not sure if there need no lock
func (w *AsynchronousWriter) writer() {
	var err error
	w.wg.Done()
	for {
		select {
		case b := <-w.queue:
			if _, err = w.file.Write(b); err != nil {
				w.errChan <- err
			}
			_asyncBufferPool.Put(b)
		case <-w.ctx:
			return
		}
	}
}

func (w *BufferWriter) Write(b []byte) (int, error) {
	var ok = false
	for ! ok {
		select {
		case filename := <-w.fire:
			if err := w.Reopen(filename); err != nil {
				return 0, err
			}
		default:
			ok = true
		}
	}

	w.lockBuf.Lock()
	*(w.buf) = append(*w.buf, b...)
	w.lockBuf.Unlock()

	if len(*w.buf) > w.cf.BufferWriterThershould && atomic.CompareAndSwapInt32(&w.swaping, 0, 1) {
		nb := make([]byte, 0, w.cf.BufferWriterThershould*2)
		ob := atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&w.buf)), (unsafe.Pointer(&nb)))
		w.file.Write(*(*[]byte)(ob))
		atomic.StoreInt32(&w.swaping, 0)
	}
	return len(b), nil
}

// Close the file and return
func (w *Writer) Close() error {
	return (*os.File)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&w.file)))).Close()
}

// Close lock and close the file
func (w *LockedWriter) Close() error {
	w.Lock()
	defer w.Unlock()
	return w.file.Close()
}

// Close set closed and close the file once
func (w *AsynchronousWriter) Close() error {
	if atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		close(w.ctx)
		w.onClose()
		return w.file.Close()
	}
	return ErrClosed
}

// onClose process remaining bufferd data for asynchronous writer
func (w *AsynchronousWriter) onClose() {
	var err error
	for {
		select {
		case b := <-w.queue:
			// flush all remaining field
			if _, err = w.file.Write(b); err != nil {
				select {
				case w.errChan <- err:
				default:
					_asyncBufferPool.Put(b)
					return
				}
			}
			_asyncBufferPool.Put(b)
		default: // after the queue was empty, return
			return
		}
	}
}

// Close bufferWriter flush all buffered write then close file
func (w *BufferWriter) Close() error {
	w.lockBuf.Lock()
	defer w.lockBuf.Unlock()
	w.file.Write(*w.buf)
	return w.file.Close()
}
