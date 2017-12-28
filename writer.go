package rollingwriter

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

// Writer provide a synchronous writer
// if Lock is set true, write will be guaranteed by lock
type Writer struct {
	file            *os.File
	absolutePath    string
	fire            chan string
	cf              *Config
	rollingfilelist chan string
}

// LockedWriter provide a synchronous writer with lock
// write operate will be guaranteed by lock
type LockedWriter struct {
	lock sync.Mutex
	Writer
}

// AsynchronousWriter provied a asynchronous writer with the writer to confirm the write
// TODO TBD if Sync is needed to be called
type AsynchronousWriter struct {
	ctx     chan int
	queue   chan []byte
	errChan chan error
	wg      sync.WaitGroup
	Writer
}

// buffer pool for asynchronous writer
var _asyncBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, BufferSize)
	},
}

// NewWriterFromConfig generate the rollingWriter with given config file
func NewWriterFromConfigFile(path string) (RollingWriter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	cfg := NewDefaultConfig()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buf, cfg)
	if err != nil {
		return nil, err
	}

	return NewWriterFromConfig(cfg)
}

// NewWriterFromConfig generate the rollingWriter with given option
func NewWriter(ops ...Option) (RollingWriter, error) {
	cfg := NewDefaultConfig()
	for _, opt := range ops {
		opt(cfg)
	}
	return NewWriterFromConfig(cfg)
}

// NewWriterFromConfig generate the rollingWriter with given config
func NewWriterFromConfig(c *Config) (RollingWriter, error) {
	// makeup log path and create
	if c.LogPath == "" || c.FileName == "" {
		return nil, ErrInvalidArgument
	}

	// make dir for path if not exist
	err := os.MkdirAll(c.LogPath, 0755)
	if err != nil {
		return nil, err
	}

	filepath := LogFilePath(c)
	// open the file and get the FD
	file, err := os.OpenFile(filepath, DefaultFileFlag, DefaultFileMode) // Open file witl rw-r--r-- by default
	if err != nil {
		return nil, err
	}

	filechan := make(chan string)
	if c.MaxRemain > 0 {
		filechan = make(chan string, c.MaxRemain+1)
	}

	mng, err := NewManager(c)
	if err != nil {
		return nil, err
	}
	var writer RollingWriter
	if c.Asynchronous { // async writer
		wr := &AsynchronousWriter{
			ctx:     make(chan int),
			queue:   make(chan []byte, QueueSize),
			errChan: make(chan error),
			wg:      sync.WaitGroup{},
			Writer: Writer{
				file:            file,
				absolutePath:    filepath,
				fire:            mng.Fire(),
				cf:              c,
				rollingfilelist: filechan,
			},
		}
		// start the asynchronous writer
		wr.wg.Add(1)
		go wr.writer()
		wr.wg.Wait()
		writer = wr
	} else {
		if c.Lock { // locked writer
			writer = &LockedWriter{
				Writer: Writer{
					file:            file,
					absolutePath:    filepath,
					fire:            mng.Fire(),
					cf:              c,
					rollingfilelist: filechan,
				},
			}
		} else { // normal writer
			writer = &Writer{
				file:            file,
				absolutePath:    filepath,
				fire:            mng.Fire(),
				cf:              c,
				rollingfilelist: filechan,
			}
		}
	}

	return writer, nil
}

// AutoRemove will delete the oldest file
func (w *Writer) AutoRemove() error {
	if len(w.rollingfilelist) > w.cf.MaxRemain {
		// remove the oldest file
		file := <-w.rollingfilelist
		return os.Remove(file)
	}
	return nil
}

// CompressFile compress log file write into .gz and remove source file
func CompressFile(oldfile *os.File, filename string) error {
	cmpname := filename + ".gz"
	cmpfile, err := os.OpenFile(cmpname, DefaultFileFlag, DefaultFileMode)
	defer cmpfile.Close()
	if err != nil {
		return err
	}
	gw := gzip.NewWriter(cmpfile)
	defer gw.Close()

	if _, err := oldfile.Seek(0, 0); err != nil {
		return err
	}
	if _, err := io.Copy(gw, oldfile); err != nil {
		cmpfile.Close()
		os.Remove(cmpname)
		return err
	}
	return os.Remove(filename) //remove *.log file
}

// Reopen is parallel safe? FIXME BUG
// func (w *LockedWriter) Reopen(file string) error {
// func (w *AsynchronousWriter) Reopen(file string) error {
func (w *Writer) Reopen(file string) error {
	// do the rename
	if err := os.Rename(w.absolutePath, file); err != nil {
		return err
	}

	// open & swap the file
	oldfile := w.file
	newfile, err := os.OpenFile(w.absolutePath, DefaultFileFlag, DefaultFileMode)
	if err != nil {
		return err
	}
	w.file = newfile

	if w.cf.Compress {
		go CompressFile(oldfile, file)
	}
	if w.cf.MaxRemain > 0 {
		go w.AutoRemove()
	}
	// BUG error handle

	return oldfile.Close()
}

func (w *Writer) Write(b []byte) (int, error) {
	select {
	case filename := <-w.fire:
		// do the reopen
		if err := w.Reopen(filename); err != nil {
			return 0, err
		}

		return w.file.Write(b)
	default:
		return w.file.Write(b)
	}
	return -1, ErrInternal
}

func (w *LockedWriter) Write(b []byte) (int, error) {
	select {
	case filename := <-w.fire:
		// do the reopen
		if err := w.Reopen(filename); err != nil {
			return 0, err
		}

		w.lock.Lock()
		defer w.lock.Unlock()
		return w.file.Write(b)
	default:
		w.lock.Lock()
		defer w.lock.Unlock()
		return w.file.Write(b)
	}
	return -1, ErrInternal
}

// Only when the error channel is empty, otherwise nothing will write and the last error will be return
func (w *AsynchronousWriter) Write(b []byte) (int, error) {
	select {
	case err := <-w.errChan:
		// return the error
		// NOTICE this error is caused by last write
		return 0, err
	case filename := <-w.fire:
		// do the reopen
		if err := w.Reopen(filename); err != nil {
			return 0, err
		}

		l := len(b)
		for len(b) > 0 {
			buf := (_asyncBufferPool.Get()).([]byte)
			n := copy(buf, b)
			// Write on the Close Channel BUG
			w.queue <- buf
			b = b[n:]
		}
		return l, nil
	default:
		// TODO TBD ???
		// is here we need not to block while the channel is full?
		// Or return a special error?
		//select {
		//	case w.queue <- b:
		//		return len(b), nil
		//	default: XXX
		//		return 0,ERRFULL
		//}

		l := len(b)
		for len(b) > 0 {
			buf := (_asyncBufferPool.Get()).([]byte)
			n := copy(buf, b)
			// Write on the Close Channel BUG
			w.queue <- buf
			b = b[n:]
		}
		return l, nil
	}
	return -1, ErrInternal
}

// writer do the asynchronous write independently
// Take care of reopen, I am not sure if there need no lock
func (w *AsynchronousWriter) writer() {
	var err error
	var b []byte
	w.wg.Done()
	for {
		select {
		case b = <-w.queue:
			if _, err = w.file.Write(b); err != nil {
				w.errChan <- err
				_asyncBufferPool.Put(b)
				// writer exit on error
				return
			}
			_asyncBufferPool.Put(b)
		case <-w.ctx:
			// writer exit on context closed
			return
		}
	}
}

// onClose process remaining bufferd data for asynchronous writer
func (w *AsynchronousWriter) onClose() error {
	var err error
	var b []byte
	for b = range w.queue {
		if _, err = w.file.Write(b); err != nil {
			_asyncBufferPool.Put(b)
			// writer exit on error
			return err
		}
		_asyncBufferPool.Put(b)
	}
	return nil
}

// Close the file and return
func (w *Writer) Close() error {
	return w.file.Close()
}

// Close lock and close the file
func (w *LockedWriter) Close() error {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.file.Close()
}

// Close lock and close the file
func (w *AsynchronousWriter) Close() error {
	close(w.queue)
	close(w.ctx)
	if err := w.onClose(); err != nil {
		w.errChan <- err
	}
	return w.file.Close()
}
