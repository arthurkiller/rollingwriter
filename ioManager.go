package rollingWriter

import (
	"strconv"
	"strings"
	"time"
)

type ioManager interface {
	Path() string
	IgnoreOK() bool
	Compress() bool
	Enable() (string, bool)
}

func NewioManager(ops ...Option) ioManager {
	return nil
}

type Option func(*patternManager)

func WithParten(pattern string) Option {
	return func(p *patternManager) {
		p.pattern = pattern
	}
}

type patternManager struct {
	filePath string
	prefix   string
	sufix    string
	// pattern is just like the crontable style
	// days-hours-minutes
	// For example:
	// 7-0-0
	// 6-23-60 means the event will fire every 7 days
	pattern       string
	rollingPoint  time.Time
	timestamp     string
	ignoreOK      bool
	rolling       bool
	compress      bool
	compressDelay bool
}

func (p *patternManager) patternUnmarshal() time.Duration {
	timeCommand := strings.Split(p.pattern, "-")
	if len(timeCommand) != 3 {
		// TODO fix the panic if needed
		panic(ErrInvalidArgument)
	}

	var offset time.Duration

	i, err := strconv.Atoi(timeCommand[0])
	if err != nil {
		panic(ErrInvalidArgument)
	}
	offset += time.Hour * 24 * time.Duration(i)

	i, err = strconv.Atoi(timeCommand[0])
	if err != nil {
		panic(ErrInvalidArgument)
	}
	offset += time.Hour * time.Duration(i)

	i, err = strconv.Atoi(timeCommand[0])
	if err != nil {
		panic(ErrInvalidArgument)
	}
	offset += time.Minute * time.Duration(i)

	return offset
}

func (p *patternManager) caculateRollingPoint() {
	p.rollingPoint = time.Now().Add(p.patternUnmarshal())
}

func (p *patternManager) Path() string {
	return p.filePath
}
func (p *patternManager) IgnoreOK() bool { return p.ignoreOK }
func (p *patternManager) Compress() bool { return p.compress }

func (p *patternManager) Enable() (string, bool) {

	now := time.Now()
	if now.Before(p.rollingPoint) {
		return "", false
	}

	p.caculateRollingPoint()
	return p.prefix + now.Format("200601021504") + p.sufix, true
}
