package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	file    *os.File
	enabled bool
)

// InitLogger initializes or disables the diagnostic logger.
func InitLogger(logPath string, enable bool) error {
	mu.Lock()
	defer mu.Unlock()

	enabled = enable
	if !enable {
		if file != nil {
			_ = file.Close()
			file = nil
		}
		return nil
	}

	if file != nil {
		_ = file.Close()
	}

	if logPath == "" {
		logPath = "launcher-debug.log"
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	file = f
	return nil
}

func logMessage(level, format string, v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if !enabled || file == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, v...)
	line := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, msg)
	_, _ = file.WriteString(line)
}

func Debug(format string, v ...interface{}) {
	logMessage("DEBUG", format, v...)
}

func Info(format string, v ...interface{}) {
	logMessage("INFO", format, v...)
}

func Error(format string, v ...interface{}) {
	logMessage("ERROR", format, v...)
}

func Close() error {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		err := file.Close()
		file = nil
		return err
	}
	return nil
}
