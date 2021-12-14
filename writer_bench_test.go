package rollingwriter

import (
	"crypto/rand"
	"io"
	"testing"
)

func BenchmarkWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	_, err := rand.Read(bf)
	if err != nil {
		return
	}

	w = newWriter()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err = w.Write(bf)
		if err != nil {
			return
		}
	}
	err = w.Close()
	if err != nil {
		return
	}
	clean()
}

func BenchmarkParallelWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	_, err := rand.Read(bf)
	if err != nil {
		return
	}

	w = newWriter()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err = w.Write(bf)
			if err != nil {
				return
			}
		}
	})
	err = w.Close()
	if err != nil {
		return
	}
	clean()
}

func BenchmarkAsynWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	_, err := rand.Read(bf)
	if err != nil {
		return
	}

	w = newAsynWriter()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err = w.Write(bf)
		if err != nil {
			return
		}
	}
	err = w.Close()
	if err != nil {
		return
	}
	clean()
}

func BenchmarkParallelAsynWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	_, err := rand.Read(bf)
	if err != nil {
		return
	}

	w = newAsynWriter()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err = w.Write(bf)
			if err != nil {
				return
			}
		}
	})
	err = w.Close()
	if err != nil {
		return
	}
	clean()
}

func BenchmarkLockedWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	_, err := rand.Read(bf)
	if err != nil {
		return
	}

	w = newLockedWriter()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := w.Write(bf)
		if err != nil {
			return
		}
	}
	err = w.Close()
	if err != nil {
		return
	}
	clean()
}

func BenchmarkParallelLockedWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	_, err := rand.Read(bf)
	if err != nil {
		return
	}

	w = newLockedWriter()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err = w.Write(bf)
			if err != nil {
				return
			}
		}
	})
	err = w.Close()
	if err != nil {
		return
	}
	clean()
}

func BenchmarkBufferWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	_, err := rand.Read(bf)
	if err != nil {
		return
	}

	w = newBufferWriter()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err = w.Write(bf)
		if err != nil {
			return
		}
	}
	err = w.Close()
	if err != nil {
		return
	}
	clean()
}

func BenchmarkParallelBufferWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	_, err := rand.Read(bf)
	if err != nil {
		return
	}

	w = newBufferWriter()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err = w.Write(bf)
			if err != nil {
				return
			}
		}
	})
	err = w.Close()
	if err != nil {
		return
	}
	clean()
}
