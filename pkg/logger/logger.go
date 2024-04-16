package logger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	timestampFormat = "2006-01-02 15:04:05Z07:00"
	red             = 31
	yellow          = 33
	blue            = 36
	gray            = 37
)

type formatter struct{}

func (m *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format(timestampFormat)
	level := getColoredLevel(entry.Level)

	var newLog string

	if entry.HasCaller() {
		fName := filepath.Base(entry.Caller.File)                    // Get the filename
		location := fmt.Sprintf("[%s:%d]", fName, entry.Caller.Line) // Get the filename and line number
		newLog = fmt.Sprintf("%s[%s] %s %s\n", level, timestamp, location, entry.Message)
	} else {
		newLog = fmt.Sprintf("%s[%s] %s\n", level, timestamp, entry.Message)
	}

	b.WriteString(newLog)
	return b.Bytes(), nil
}

func getColoredLevel(level logrus.Level) string {
	levelStr := strings.ToUpper(level.String())
	var color int
	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:

		color = gray
	case logrus.WarnLevel:
		color = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		color = red
	case logrus.InfoLevel:
		color = blue
	default:
		color = blue
	}

	return fmt.Sprintf("\x1b[%dm%s\u001B[0m", color, levelStr[:4])
}

func SetLogger() {
	logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&formatter{})
}
