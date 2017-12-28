package rollingwriter

import "testing"

func BenchmarkWrite(b *testing.B) {
	// 原因：并发创建了多个filedescripter ,并发读写导致错误
	w := newTWriter()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w.Write([]byte("hello,world"))
	}

	w = newVWriter()
	for i := 0; i < b.N; i++ {
		w.Write([]byte("hello,world"))
	}
}

func BenchmarkParallelWrite(b *testing.B) {
	w := newTWriter()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write([]byte("hello,world"))
		}
	})

	w = newVWriter()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write([]byte("hello,world"))
		}
	})
}
