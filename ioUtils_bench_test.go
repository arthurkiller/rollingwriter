package bunnystub

import "testing"

func BenchmarkWrite(b *testing.B) {
	// 原因：并发创建了多个filedescripter ,并发读写导致错误
	w := newWriter()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w.Write([]byte("hello,world"))
	}
}

func BenchmarkParallelWrite(b *testing.B) {
	w := newWriter()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Write([]byte("hello,world"))
		}
	})
}
