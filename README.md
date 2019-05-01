# RollingWriter [![Build Status](https://travis-ci.org/arthurkiller/rollingWriter.svg?branch=master)](https://travis-ci.org/arthurkiller/rollingWriter) [![Go Report Card](https://goreportcard.com/badge/github.com/arthurkiller/rollingwriter)](https://goreportcard.com/report/github.com/arthurkiller/rollingwriter) [![GoDoc](https://godoc.org/github.com/arthurkiller/rollingWriter?status.svg)](https://godoc.org/github.com/arthurkiller/rollingWriter) [![codecov](https://codecov.io/gh/arthurkiller/rollingwriter/branch/master/graph/badge.svg)](https://codecov.io/gh/arthurkiller/rollingwriter)
RollingWriter is an auto rotate `io.Writer` implementation. It can works well with logger.

__New Version v2.0 is comming out! Much more Powerfull and Efficient. Try it by follow the demo__

it contains 2 separate patrs:
* Manager: decide when to rotate the file with policy
    RlingPolicy give out the rolling policy
    * WithoutRolling: no rolling will happen
    * TimeRolling: rolling by time
    * VolumeRolling: rolling by file size

* IOWriter: impement the io.Writer and do the io write
    * Writer: not parallel safe writer
    * AsyncWtiter: parallel safe async writer
    * BufferWriter: merge serval write into one `file.Write()`

## Features
* Auto rotate with multi rotate policies
* Implement go io.Writer, provide parallel safe writer
* Max remain rolling files with auto cleanup
* Multi writer mode

## Benchmark
```bash
goos: darwin
goarch: amd64
pkg: github.com/arthurkiller/rollingWriter
BenchmarkWrite-4                          200000              8400 ns/op               0 B/op          0 allocs/op
BenchmarkParallelWrite-4                  100000             12944 ns/op               0 B/op          0 allocs/op
BenchmarkAsynWrite-4                      100000             12444 ns/op           40356 B/op          1 allocs/op
BenchmarkParallelAsynWrite-4               10000            105036 ns/op          245173 B/op          1 allocs/op
BenchmarkBufferWrite-4                    100000             11697 ns/op            1984 B/op          4 allocs/op
BenchmarkParallelBufferWrite-4           1000000              1142 ns/op            7752 B/op          1 allocs/op
PASS
ok      github.com/arthurkiller/rollingWriter   9.547s
```

## Quick Start
```golang
	writer, err := rollingwriter.NewWriterFromConfig(&config)
	if err != nil {
		panic(err)
	}

	writer.Write([]byte("hello, world"))
```
Want more? View `demo` for more details.

Any suggestion or new feature inneed, please [put up an issue](https://github.com/arthurkiller/rollingWriter/issues/new)
