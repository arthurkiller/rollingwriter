package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/arthurkiller/rollingwriter"
)

func main() {
	// writer 实现了 io.Writer 的全部接口
	// 使用配置方式生成一个 writer 或者 Option 都可以
	config := rollingwriter.Config{
		LogPath:       "./log",               //日志路径
		TimeTagFormat: "060102150405",        //时间格式串 如加入则会在文件的末尾加上格式化后的时间字符串 如：xxx.log.20210202
		FileName:      "test-%Y-%m-%d-%H-%M", //日志文件名 支持格式化时间 更多使用方式请查阅 https://github.com/lestrrat-go/strftime

		// 可同时设置，两种限制优先达到哪一个优先选用哪一个
		MaxRemain: 5,             //配置日志最大存留数
		MaxAge:    5 * time.Hour, //日志最长留存时间

		// 目前有2中滚动策略: 按照时间滚动按照大小滚动
		// - 时间滚动: 配置策略如同 crontable, 例如,每天0:0切分, 则配置 0 0 0 * * *
		// - 大小滚动: 配置单个日志文件(未压缩)的滚动大小门限, 如1G, 500M
		RollingPolicy:      rollingwriter.TimeRolling, //配置滚动策略 norolling timerolling volumerolling
		RollingTimePattern: "* * * * * *",             //配置时间滚动策略
		RollingVolumeSize:  "2k",                      //配置截断文件下限大小

		// writer 支持4种不同的 mode:
		// 1. none 2. lock
		// 3. async 4. buffer
		// - 无保护的 writer: 不提供并发安全保障
		// - lock 保护的 writer: 提供由 mutex 保护的并发安全保障
		// - 异步 writer: 异步 write, 并发安全. 异步开启后忽略 Lock 选项
		WriterMode: "lock",
		// BufferWriterThershould in B
		BufferWriterThershould: 8 * 1024 * 1024,
		// Compress will compress log file with gzip
		Compress: true,
	}

	// 创建一个 writer
	writer, err := rollingwriter.NewWriterFromConfig(&config)
	if err != nil {
		// 应该处理错误
		panic(err)
	}

	// 并发读写即可
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for {
				fmt.Fprintf(writer, "now :%s \n", time.Now())
			}
		}()
	}
	wg.Wait()
}
