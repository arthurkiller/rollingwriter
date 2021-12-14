package rollingwriter

import (
	"github.com/lestrrat-go/strftime"
	"path"
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
	c.RollingVolumeSize = "1tb"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024*1024*1024*1024), m.thresholdSize)
	c.RollingVolumeSize = "1tB"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024*1024*1024*1024), m.thresholdSize)
	c.RollingVolumeSize = "1t"
	m.ParseVolume(c)
	assert.Equal(t, int64(1024*1024*1024*1024), m.thresholdSize)
}

func TestGenLogFileName(t *testing.T) {
	pattern, _ := strftime.New("file")
	m := manager{
		LogPath:       "./",
		FileName:      pattern,
		TimeTagFormat: "200601021504",
	}
	m.startAt = time.Now()

	dest := m.GenLogFileName()
	timetag := m.startAt.Format(m.TimeTagFormat)
	assert.Equal(t, path.Join("./", "file"+".log."+timetag), dest)
}
