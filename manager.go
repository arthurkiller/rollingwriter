package rollingwriter

import (
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron"
)

type manager struct {
	thresholdSize int64
	startAt       time.Time
	fire          chan string
	cr            *cron.Cron
	context       chan int
	wg            sync.WaitGroup
	lock          sync.Mutex
}

// NewManager generate the Manager with config
func NewManager(c *Config) (Manager, error) {
	m := &manager{
		startAt: time.Now(),
		cr:      cron.New(),
		fire:    make(chan string),
		context: make(chan int),
		wg:      sync.WaitGroup{},
	}

	// start the manager according to policy
	switch c.RollingPolicy {
	default:
		fallthrough
	case WithoutRolling:
		return m, nil
	case TimeRolling:
		if err := m.cr.AddFunc(c.RollingTimePattern, func() {
			m.fire <- m.GenLogFileName(c)
		}); err != nil {
			return nil, err
		}
		m.cr.Start()
	case VolumeRolling:
		m.ParseVolume(c)
		m.wg.Add(1)
		go func() {
			timer := time.NewTicker(time.Duration(Precision) * time.Second)
			defer timer.Stop()

			filepath := LogFilePath(c)
			var file *os.File
			var err error
			m.wg.Done()

			for {
				select {
				case <-m.context:
					return
				case <-timer.C:
					if file, err = os.Open(filepath); err != nil {
						continue
					}
					if info, err := file.Stat(); err == nil && info.Size() > m.thresholdSize {
						m.fire <- m.GenLogFileName(c)
					}
					file.Close()
				}
			}
		}()
		m.wg.Wait()
	}
	return m, nil
}

// Fire return the fire channel
func (m *manager) Fire() chan string {
	return m.fire
}

// Close return stop the manager and return
func (m *manager) Close() {
	close(m.context)
	m.cr.Stop()
}

// ParseVolume parse the config volume format and return threshold
func (m *manager) ParseVolume(c *Config) {
	s := []byte(strings.ToUpper(c.RollingVolumeSize))
	if !(strings.Contains(string(s), "K") || strings.Contains(string(s), "KB") ||
		strings.Contains(string(s), "M") || strings.Contains(string(s), "MB") ||
		strings.Contains(string(s), "G") || strings.Contains(string(s), "GB") ||
		strings.Contains(string(s), "T") || strings.Contains(string(s), "TB")) {

		// set the default threshold with 1GB
		m.thresholdSize = 1024 * 1024 * 1024
		return
	}

	var unit int64 = 1
	p, _ := strconv.Atoi(string(s[:len(s)-1]))
	unitstr := string(s[len(s)-1])

	if s[len(s)-1] == 'B' {
		p, _ = strconv.Atoi(string(s[:len(s)-2]))
		unitstr = string(s[len(s)-2:])
	}

	switch unitstr {
	default:
		fallthrough
	case "T", "TB":
		unit *= 1024
		fallthrough
	case "G", "GB":
		unit *= 1024
		fallthrough
	case "M", "MB":
		unit *= 1024
		fallthrough
	case "K", "KB":
		unit *= 1024
	}
	m.thresholdSize = int64(p) * unit
}

// GenLogFileName generate the new log file name, filename should be absolute path
func (m *manager) GenLogFileName(c *Config) (filename string) {
	m.lock.Lock()
	// [path-to-log]/filename.log.2007010215041517
	if c.Compress {
		filename = path.Join(c.LogPath, c.FileName+".log.gz."+m.startAt.Format(c.TimeTagFormat))
	} else {
		filename = path.Join(c.LogPath, c.FileName+".log."+m.startAt.Format(c.TimeTagFormat))
	}
	// reset the start time to now
	m.startAt = time.Now()
	m.lock.Unlock()
	return
}
