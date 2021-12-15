package rollingwriter

import (
	"io"
	"math/rand"
	"os"
	"testing"
)

func clean() {
	_ = os.Remove("./test/unittest.log")
	_ = os.Remove("./test/unittest.reopen")
	_ = os.Remove("./test/unittest.gz")
	_ = os.Remove("./test")
}

func newWriter() *Writer {
	cfg := NewDefaultConfig()
	cfg.LogPath = "./test"
	cfg.FileName = "unittest"
	cfg.WriterMode = "none"
	w, _ := NewWriterFromConfig(&cfg)
	return w.(*Writer)
}

func newVolumeWriter() *Writer {
	cfg := NewDefaultConfig()
	cfg.RollingPolicy = 3
	cfg.RollingVolumeSize = "1mb"
	cfg.LogPath = "./test"
	cfg.FileName = "unittest"
	cfg.WriterMode = "none"
	w, _ := NewWriterFromConfig(&cfg)
	return w.(*Writer)
}

func newLockedWriter() *LockedWriter {
	cfg := NewDefaultConfig()
	cfg.LogPath = "./test"
	cfg.FileName = "unittest"
	cfg.WriterMode = "lock"
	w, _ := NewWriterFromConfig(&cfg)
	return w.(*LockedWriter)
}

func newAsynWriter() *AsynchronousWriter {
	cfg := NewDefaultConfig()
	cfg.LogPath = "./test"
	cfg.FileName = "unittest"
	cfg.WriterMode = "async"
	w, _ := NewWriterFromConfig(&cfg)
	return w.(*AsynchronousWriter)
}

func newBufferWriter() *BufferWriter {
	cfg := NewDefaultConfig()
	cfg.LogPath = "./test"
	cfg.FileName = "unittest"
	cfg.WriterMode = "buffer"
	w, _ := NewWriterFromConfig(&cfg)
	return w.(*BufferWriter)
}

func TestNewWriter(t *testing.T) {
	if _, err := NewWriter(
		WithTimeTagFormat("200601021504"), WithLogPath("./"), WithFileName("foo"),
		WithAsynchronous(), WithBuffer(), WithBufferThershould(8), WithCompress(), WithLock(),
		WithMaxRemain(3), WithRollingVolumeSize("100mb"), WithRollingTimePattern("0 0 0 * * *"),
	); err != nil {
		t.Fatal("error in test new writer", err)
	}
	_ = os.Remove("./foo.log")
}

func TestWrite(t *testing.T) {
	var writer io.WriteCloser
	var c int = 1000
	var l int = 1024

	writer = newWriter()
	for i := 0; i < c; i++ {
		bf := make([]byte, l)
		rand.Read(bf)
		_, _ = writer.Write(bf)
	}
	_ = writer.Close()
	clean()

	writer = newVolumeWriter()
	for i := 0; i < c; i++ {
		bf := make([]byte, l)
		rand.Read(bf)
		_, _ = writer.Write(bf)
	}
	_ = writer.Close()
	clean()

	writer = newLockedWriter()
	for i := 0; i < c; i++ {
		bf := make([]byte, l)
		rand.Read(bf)
		_, _ = writer.Write(bf)
	}
	_ = writer.Close()
	clean()

	writer = newAsynWriter()
	for i := 0; i < c; i++ {
		bf := make([]byte, l)
		rand.Read(bf)
		_, _ = writer.Write(bf)
	}
	_ = writer.Close()
	clean()

	writer = newBufferWriter()
	for i := 0; i < c; i++ {
		bf := make([]byte, l)
		rand.Read(bf)
		_, _ = writer.Write(bf)
	}
	_ = writer.Close()
	clean()
}

func TestWriteParallel(t *testing.T) {
	var writer io.WriteCloser
	var c int = 1000
	var l int = 1024

	t.Run("none", func(t *testing.T) {
		t.Parallel()
		writer = newWriter()
		bf := make([]byte, l)
		rand.Read(bf)
		_, _ = writer.Write(bf)
		for i := 0; i < c; i++ {
			bf := make([]byte, l)
			rand.Read(bf)
			_, _ = writer.Write(bf)
		}
		_ = writer.Close()
		clean()
	})
}

func TestVolumeWriteParallel(t *testing.T) {
	var writer io.WriteCloser
	var c int = 1000
	var l int = 1024

	t.Run("none", func(t *testing.T) {
		t.Parallel()
		writer = newVolumeWriter()
		bf := make([]byte, l)
		rand.Read(bf)
		_, _ = writer.Write(bf)
		for i := 0; i < c; i++ {
			bf := make([]byte, l)
			rand.Read(bf)
			_, _ = writer.Write(bf)
		}
		_ = writer.Close()
		clean()
	})
}
func TestWriteLockParallel(t *testing.T) {
	var writer io.WriteCloser
	var c int = 1000
	var l int = 1024

	t.Run("locked", func(t *testing.T) {
		t.Parallel()
		writer = newLockedWriter()
		for i := 0; i < c; i++ {
			bf := make([]byte, l)
			rand.Read(bf)
			_, _ = writer.Write(bf)
		}
		_ = writer.Close()
		clean()
	})

}

func TestWriteAsyncParallel(t *testing.T) {
	var writer io.WriteCloser
	var c int = 1000
	var l int = 1024

	t.Run("async", func(t *testing.T) {
		t.Parallel()
		writer = newAsynWriter()
		for i := 0; i < c; i++ {
			bf := make([]byte, l)
			rand.Read(bf)
			_, _ = writer.Write(bf)
		}
		_ = writer.Close()
		clean()
	})
}

func TestWriteBufferParallel(t *testing.T) {
	var writer io.WriteCloser
	var c int = 1000
	var l int = 1024

	t.Run("buffer", func(t *testing.T) {
		t.Parallel()
		writer = newBufferWriter()
		for i := 0; i < c; i++ {
			bf := make([]byte, l)
			rand.Read(bf)
			_, _ = writer.Write(bf)
		}
		_ = writer.Close()
		clean()
	})
}

func TestReopen(t *testing.T) {
	var c int = 1000
	var l int = 1024

	t.Run("none", func(t *testing.T) {
		t.Parallel()
		writer := newWriter()
		for i := 0; i < c; i++ {
			bf := make([]byte, l)
			rand.Read(bf)
			_, _ = writer.Write(bf)
		}
		_ = writer.Reopen("./test/unittest.reopen")
		_ = writer.Close()
		clean()
	})
}

func TestAutoRemove(t *testing.T) {
	var c int = 1000
	var l int = 1024

	writer := newWriter()
	for i := 0; i < c; i++ {
		bf := make([]byte, l)
		rand.Read(bf)
		_, _ = writer.Write(bf)
	}
	_ = writer.Close()
	writer.cf.MaxRemain = 0
	clean()
}

func TestCompress(t *testing.T) {
	var c int = 1000
	var l int = 1024

	writer := newWriter()
	for i := 0; i < c; i++ {
		bf := make([]byte, l)
		rand.Read(bf)
		_, _ = writer.Write(bf)
	}
	_ = writer.CompressFile(writer.file, "./test/unittest.gz")
	_ = writer.Close()
	clean()
}
