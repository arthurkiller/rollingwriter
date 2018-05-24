package main

import (
	"crypto/rand"
	"sync"

	"github.com/arthurkiller/rollingWriter"
)

func main() {
	// writer 实现了 io.Writer 的全部接口
	// 使用配置方式生成一个 writer 或者 Option 都可以
	config := rollingwriter.Config{
		LogPath:       "./log",        //日志路径
		TimeTagFormat: "060102150405", //时间格式串
		FileName:      "test",         //日志文件名
		MaxRemain:     5,              //配置日志最大存留数

		// 目前有2中滚动策略: 按照时间滚动按照大小滚动
		// - 时间滚动: 配置策略如同 crontable, 例如,每天0:0切分, 则配置 0 0 0 * * *
		// - 大小滚动: 配置单个日志文件(未压缩)的滚动大小门限, 入1G, 500M
		RollingPolicy:      rollingwriter.TimeRolling, //配置滚动策略 norolling timerolling volumerolling
		RollingTimePattern: "1 * * * * *",             //配置时间滚动策略
		RollingVolumeSize:  "20M",                     //配置截断文件下限大小
		Compress:           true,                      //配置是否压缩存储

		// writer 支持3种方式:
		// - 无保护的 writer: 不提供并发安全保障
		// - lock 保护的 writer: 提供由 mutex 保护的并发安全保障
		// - 异步 writer: 异步 write, 并发安全. 异步开启后忽略 Lock 选项
		Asynchronous: true, //配置是否异步写
		Lock:         true, //配置是否同步加锁写
	}

	// 创建一个 writer
	writer, err := rollingwriter.NewWriterFromConfig(&config)
	if err != nil {
		// 应该处理错误
		panic(err)
	}

	// 并发读写即可
	wg := sync.WaitGroup{}
	bf := make([]byte, 128)
	rand.Read(bf)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			for {
				writer.Write(bf)
			}
		}()
	}

	wg.Wait()
}
