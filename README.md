# RollingWriter [![Build Status](https://travis-ci.org/arthurkiller/rollingWriter.svg?branch=master)](https://travis-ci.org/arthurkiller/rollingWriter) [![Go Report Card](https://goreportcard.com/badge/github.com/arthurkiller/rollingwriter)](https://goreportcard.com/report/github.com/arthurkiller/rollingwriter) [![GoDoc](https://godoc.org/github.com/arthurkiller/rollingWriter?status.svg)](https://godoc.org/github.com/arthurkiller/rollingWriter) [![codecov](https://codecov.io/gh/arthurkiller/rollingwriter/branch/master/graph/badge.svg)](https://codecov.io/gh/arthurkiller/rollingwriter)[![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go#logging)
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
    * LockedWriter: parallel safe garented by lock
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
BenchmarkWrite-4                          300000              5952 ns/op               0 B/op          0 allocs/op
BenchmarkParallelWrite-4                  200000              7846 ns/op               0 B/op          0 allocs/op
BenchmarkAsynWrite-4                      200000              7917 ns/op           16324 B/op          1 allocs/op
BenchmarkParallelAsynWrite-4              200000              8632 ns/op           12513 B/op          1 allocs/op
BenchmarkLockedWrite-4                    200000              5829 ns/op               0 B/op          0 allocs/op
BenchmarkParallelLockedWrite-4            200000              7796 ns/op               0 B/op          0 allocs/op
BenchmarkBufferWrite-4                    200000              6943 ns/op            1984 B/op          4 allocs/op
BenchmarkParallelBufferWrite-4           1000000              1026 ns/op            7129 B/op          1 allocs/op
PASS
ok      github.com/arthurkiller/rollingWriter   14.867s
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
