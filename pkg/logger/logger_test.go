package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestTC_LOG_01_ThreadSafeLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test-debug.log")

	if err := InitLogger(logPath, true); err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer func() { _ = Close() }()

	var wg sync.WaitGroup
	workers := 10
	linesPerWorker := 100

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < linesPerWorker; j++ {
				Info("Worker %d message %d", workerID, j)
			}
		}(i)
	}

	wg.Wait()
	_ = Close()

	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	totalLines := 0
	for scanner.Scan() {
		totalLines++
		line := scanner.Text()
		if !strings.Contains(line, "[INFO]") || !strings.Contains(line, "Worker") {
			t.Errorf("Malformed log line: %s", line)
		}
	}

	expected := workers * linesPerWorker
	if totalLines != expected {
		t.Errorf("Expected %d log lines, found %d", expected, totalLines)
	}
}
