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
	rand.Read(bf)

	w = newWriter()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Write(bf)
	}
	w.Close()
	clean()
}

func BenchmarkParallelWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	rand.Read(bf)

	w = newWriter()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write(bf)
		}
	})
	w.Close()
	clean()
}

func BenchmarkAsynWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	rand.Read(bf)

	w = newAsynWriter()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Write(bf)
	}
	w.Close()
	clean()
}

func BenchmarkParallelAsynWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	rand.Read(bf)

	w = newAsynWriter()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write(bf)
		}
	})
	w.Close()
	clean()
}

func BenchmarkLockedWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	rand.Read(bf)

	w = newLockedWriter()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Write(bf)
	}
	w.Close()
	clean()
}

func BenchmarkParallelLockedWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	rand.Read(bf)

	w = newLockedWriter()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write(bf)
		}
	})
	w.Close()
	clean()
}

func BenchmarkBufferWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	rand.Read(bf)

	w = newBufferWriter()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Write(bf)
	}
	w.Close()
	clean()
}

func BenchmarkParallelBufferWrite(b *testing.B) {
	var w io.WriteCloser
	var l int = 1024
	bf := make([]byte, l)
	rand.Read(bf)

	w = newBufferWriter()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write(bf)
		}
	})
	w.Close()
	clean()
}
