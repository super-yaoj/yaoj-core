// log utilties
package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Entry = logrus.Entry

type Fields = logrus.Fields

// log to terminal
func NewTerminal() *Entry {
	return logrus.NewEntry(&logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			ForceColors: true,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.InfoLevel,
	})
}

// 测试时使用的 logger （彩色、DebugLevel）
func NewTest() *Entry {
	return logrus.NewEntry(&logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			ForceColors: true,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.DebugLevel,
	})
}
