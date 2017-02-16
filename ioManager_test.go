package bunnystub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newioManager() *patternManager {
	return &patternManager{
		pattern:  "1-2-3",
		compress: false,
	}
}

func TestPatternUnmarshal(t *testing.T) {
	n := newioManager()
	n.pattern = "1-2-3"
	offset := n.patternUnmarshal()
	d := time.Duration(time.Hour*26 + time.Minute*3)

	assert.Equal(t, d, offset)
}

func TestCaculateRollingPoint(t *testing.T) {
	n := newioManager()
	n.pattern = "1-2-3"
	d := time.Duration(time.Hour*26 + time.Minute*3)
	destrp := time.Now().Add(d)
	n.caculateRollingPoint(time.Now())

	assert.Equal(t, destrp.Format("200601021504"), n.rollingPoint.Format("200601021504"))
}

func TestEnable(t *testing.T) {
	n := newioManager()
	n.pattern = "1-2-3"
	n.rollingPoint = time.Now().Add(-time.Minute)
	n.prefix = "@head"
	n.suffix = "@end"

	s, ok := n.Enable()
	d := time.Duration(time.Hour*26 + time.Minute*3)

	assert.Equal(t, true, ok)

	assert.Equal(t, "@head"+time.Now().Add(-d).Format("200601021504")+"@end.log", s)

	destrp := time.Now().Add(d)
	assert.Equal(t, destrp.Format("200601021504"), n.rollingPoint.Format("200601021504"))
}
