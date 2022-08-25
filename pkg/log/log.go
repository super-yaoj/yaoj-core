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
		Level: logrus.DebugLevel,
	})
}
