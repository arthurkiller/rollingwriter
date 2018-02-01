package bunnystub

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTimeManager() manager {
	return manager{}
}
func newSizeManager() manager {
	return manager{}
}

func TestParseVolume(t *testing.T) {
	c := &Config{}
	m := manager{}

	c.RollingVolumeSize = "1kb"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024), m.thresholdSize)
	c.RollingVolumeSize = "2k"
	m.ParseVolume(c)
	assert.Equal(t, int64(2*1024), m.thresholdSize)
	c.RollingVolumeSize = "1KB"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024), m.thresholdSize)
	c.RollingVolumeSize = "1mb"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024*1024), m.thresholdSize)
	c.RollingVolumeSize = "1MB"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024*1024), m.thresholdSize)
	c.RollingVolumeSize = "1Mb"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024*1024), m.thresholdSize)
	c.RollingVolumeSize = "1gb"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024*1024*1024), m.thresholdSize)
	c.RollingVolumeSize = "1GB"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024*1024*1024), m.thresholdSize)
	c.RollingVolumeSize = "1g"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024*1024*1024), m.thresholdSize)
}

func TestGenLogFileName(t *testing.T) {
	c := &Config{
		Prefix:        "prefix",
		FileName:      "file",
		Suffix:        "suffix",
		Separator:     "-",
		TimeTagFormat: "200601021504",
		//		LogPath:       "./",
	}
	m := manager{}
	m.startAt = time.Now()

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

func TestFire(t *testing.T) {
}
