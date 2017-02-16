package bunnystub

import (
	"log"
	"strconv"
	"strings"
	"time"
)

type ioManager interface {
	Path() string
	IgnoreOK() bool
	Compress() bool
	Enable() (string, bool)
	NameParts() (string, string, string)
}

func NewioManager(ops ...Option) ioManager {
	m := &patternManager{
		filePath: "./",
		prefix:   "",
		suffix:   "",
		pattern:  "1-0-0",
	}

	for _, o := range ops {
		o(m)
	}

	m.caculateRollingPoint(time.Now())

	return m
}

type Option func(*patternManager)

func WithPattern(pattern string) Option {
	return func(p *patternManager) {
		p.pattern = pattern
	}
}

func WithPrefix(prefix string) Option {
	return func(p *patternManager) {
		p.prefix = prefix
	}
}
func WithSuffix(suffix string) Option {
	return func(p *patternManager) {
		p.suffix = suffix
	}
}
func WithIgnoreOK() Option {
	return func(p *patternManager) {
		p.ignoreOK = true
	}
}
func WithCompress() Option {
	return func(p *patternManager) {
		p.compress = true
	}
}

type patternManager struct {
	filePath string
	// file name js like this style: prefix-timestamp-suffix.log
	// compressed log file is named like this: prefix-timestamp-suffix.tar
	prefix string
	suffix string
	// pattern is just like the crontable style
	// days-hours-minutes
	// For example:
	// 7-0-0
	// 6-23-60 means the event will fire every 7 days
	pattern      string
	rollingPoint time.Time
	ignoreOK     bool
	compress     bool
}

func (p *patternManager) patternUnmarshal() time.Duration {
	timeCommand := strings.Split(p.pattern, "-")
	// TODO fix the return for the split
	if len(timeCommand) != 3 {
		log.Println("Invalid arguments")
		return time.Duration(time.Hour * 24)
	}

	var offset time.Duration

	i, err := strconv.Atoi(timeCommand[0])
	if err != nil {
		log.Println("Invalid arguments")
		return time.Duration(time.Hour * 24)
	}
	offset += time.Hour * time.Duration(24*i)

	i, err = strconv.Atoi(timeCommand[1])
	if err != nil {
		log.Println("Invalid arguments")
		return time.Duration(time.Hour * 24)
	}
	offset += time.Hour * time.Duration(i)

	i, err = strconv.Atoi(timeCommand[2])
	if err != nil {
		log.Println("Invalid arguments")
		return time.Duration(time.Hour * 24)
	}
	offset += time.Minute * time.Duration(i)

	return offset
}

func (p *patternManager) caculateRollingPoint(t time.Time) {
	p.rollingPoint = t.Add(p.patternUnmarshal())
}

func (p *patternManager) Enable() (string, bool) {
	now := time.Now()
	if now.Before(p.rollingPoint) {
		return "", false
	}

	dur := -p.patternUnmarshal()

	p.caculateRollingPoint(now)
	return p.prefix + now.Add(dur).Format("200601021504") + p.suffix + ".log", true
}

func (p *patternManager) Path() string {
	return p.filePath + p.prefix + p.rollingPoint.Format("200601021504") + p.suffix + ".log"
}
func (p *patternManager) IgnoreOK() bool                      { return p.ignoreOK }
func (p *patternManager) Compress() bool                      { return p.compress }
func (p *patternManager) NameParts() (string, string, string) { return p.filePath, p.prefix, p.suffix }
