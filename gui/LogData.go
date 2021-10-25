package gui

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type LogData struct {
	logger   *log.Logger
	err      error
	fileName string
	active   bool
	warning  bool
}

func NewLogData(fileName string, prefix string, active bool) *LogData {
	if fileName == "" {
		err := fmt.Errorf("log is active but log file name was not provided")
		return &LogData{active: active, logger: nil, err: err, warning: active}
	}
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return &LogData{active: active, logger: nil, err: err, warning: true}
	}
	l := log.New(file, prefix, log.Ldate|log.Ltime)
	return &LogData{active: active, fileName: fileName, logger: l, err: nil, warning: false}
}

func (lw *LogData) IsWarning() bool {
	return lw.warning
}

func (lw *LogData) GetErr() error {
	return lw.err
}

func (lw *LogData) FlipOnOff() {
	if lw.IsReady() {
		lw.active = !lw.active
	}
}

func (lw *LogData) Start() {
	if lw.IsReady() {
		lw.active = true
	}
}

func (lw *LogData) Stop() {
	if lw.IsReady() {
		lw.active = false
	}
}

func (lw *LogData) IsLogging() bool {
	if lw.IsReady() {
		return lw.active
	}
	return false
}

func (lw *LogData) IsReady() bool {
	return lw.logger != nil
}

func (lw *LogData) Close() {
}

func (lw *LogData) Log(l string) {
	if lw.IsLogging() {
		go lw.logger.Println(l)
	}
}

func LogCleanString(s string, max int) string {
	var sb strings.Builder
	count := 0
	for _, r := range s {
		if r < 32 {
			sb.WriteString(fmt.Sprintf("[%d]", r))
		} else {
			sb.WriteRune(r)
		}
		if count >= max {
			break
		}
	}
	return sb.String()
}
