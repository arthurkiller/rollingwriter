package bunnystub

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/robfig/cron"
	"github.com/stretchr/testify/assert"
)

func newTimeManager() *IOManager {
	m := &IOManager{
		FileMod:            os.FileMode(0644),
		RollingTimePattern: "* * * * * *",
		RollingVolumeSize:  "1G",
		TimeFormatPattern:  "20060102",
		enable:             make(chan string),
		IsLockFree:         false,
		IsIgnoreOK:         false,
		cr:                 cron.New(),
	}
	m.rollingPoint = time.Now()
	m.trigger = func() {}
	m.cr.AddFunc(m.RollingTimePattern, func() {
		m.enable <- m.FilePath + m.Prefix + m.Suffix + "." + m.rollingPoint.Format(m.TimeFormatPattern)
		m.rollingPoint = time.Now()
	})
	m.cr.Start()
	return m
}

//
//func newVolumeManager() *IOManager {
//	m := &IOManager{
//		FileMod:            os.FileMode(0644),
//		RollingTimePattern: "0 0 * * * *",
//		RollingVolumeSize:  "1b",
//		TimeFormatPattern:  "20060102",
//		enable:             make(chan string),
//		IsLockFree:         false,
//		IsIgnoreOK:         false,
//		cr:                 cron.New(),
//	}
//	m.rollingPoint = time.Now()
//	m.rollingVolume = 1 // 1byte
//	m.trigger = func() {
//		if atomic.LoadInt64(&m.fileSize) > m.rollingVolume {
//			m.enable <- m.FilePath + m.Prefix + m.Suffix + "." + m.rollingPoint.Format(m.TimeFormatPattern)
//			m.rollingPoint = time.Now()
//		}
//	}
//	return m
//}

func TestParseVolume(t *testing.T) {
	c := &Config{}
	m := manager{}

	c.RollingVolumeSize = "1kb"
	m.parseVolume(c)
	assert.Equal(t, int64(1024), m.rollingVolume)
	c.RollingVolumeSize = "2k"
	m.parseVolume()
	assert.Equal(t, int64(2*1024), m.rollingVolume)
	c.RollingVolumeSize = "1KB"
	m.parseVolume()
	assert.Equal(t, int64(1024), m.rollingVolume)
	c.RollingVolumeSize = "1mb"
	m.parseVolume()
	assert.Equal(t, int64(1024*1024), m.rollingVolume)
	c.RollingVolumeSize = "1MB"
	m.parseVolume()
	assert.Equal(t, int64(1024*1024), m.rollingVolume)
	c.RollingVolumeSize = "1Mb"
	m.parseVolume()
	assert.Equal(t, int64(1024*1024), m.rollingVolume)
	c.RollingVolumeSize = "1gb"
	m.parseVolume()
	assert.Equal(t, int64(1024*1024*1024), m.rollingVolume)
	c.RollingVolumeSize = "1GB"
	m.parseVolume()
	assert.Equal(t, int64(1024*1024*1024), m.rollingVolume)
	c.RollingVolumeSize = "1g"
	m.parseVolume()
	assert.Equal(t, int64(1024*1024*1024), m.rollingVolume)
}

func TestGenLogFile(t *testing.T) {
	c := &Config{}
	c.Prefix = "prefix"
	c.FileName = "file"
	c.Suffix = "suffix"
	c.Separator = "-"
	c.TimeTagFormat = "200601021504"
	m := manager{}
	m.startAt = time.Time

	dest := m.GenLogFileName(c)
	patrs := []string{"prefix", "file", "suffix"}
	timetag := m.startAt.Format(c.TimeTagFormat)
	assert.Equal(t, strings.Join(patrs, "-")+".log"+c.Separator+timetag, dest)

	c.Separator = "*"
	dest = m.GenLogFileName(c)
	patrs = []string{"prefix", "file", "suffix"}
	assert.Equal(t, strings.Join(patrs, "*")+".log"+c.Separator+timetag, dest)

	c.Prefix = ""
	dest = m.GenLogFileName(c)
	patrs = []string{"file", "suffix"}
	assert.Equal(t, strings.Join(patrs, "*")+".log"+c.Separator+timetag, dest)
}

//func TestEnable(t *testing.T) {
//	n := newTimeManager()
//
//	ch := n.Enable()
//	d := time.Duration(time.Hour*26 + time.Minute*3)
//
//	s := <-ch
//
//	assert.Equal(t, "."+time.Now().Add(-d).Format("20060102"), s)
//
//	destrp := time.Now().Add(d)
//	assert.Equal(t, destrp.Format("200601021504"), n.rollingPoint.Format("200601021504"))
//}
