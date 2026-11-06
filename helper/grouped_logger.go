package helper

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// GroupLogger writes a complete group of related log lines atomically.
// Implementations can send the group to a terminal today and to another
// destination in the future without changing a creature service.
type GroupLogger interface {
	Log(group string, messages ...string) error
}

// GroupedLogger keeps concurrent output readable by preventing related lines
// written by different goroutines from being interleaved.
type GroupedLogger struct {
	mutex  sync.Mutex
	writer io.Writer
}

func NewGroupedLogger(writer io.Writer) *GroupedLogger {
	return &GroupedLogger{writer: writer}
}

func (logger *GroupedLogger) Log(group string, messages ...string) error {
	if logger == nil || logger.writer == nil {
		return fmt.Errorf("grouped logger is not initialized")
	}

	group = strings.TrimSpace(group)
	if group == "" {
		group = "criatura"
	}

	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	if _, err := fmt.Fprintf(logger.writer, "\n[%s]\n", group); err != nil {
		return fmt.Errorf("write log group: %w", err)
	}

	for _, message := range messages {
		for _, line := range strings.Split(message, "\n") {
			if _, err := fmt.Fprintf(logger.writer, "  %s\n", line); err != nil {
				return fmt.Errorf("write log message: %w", err)
			}
		}
	}

	return nil
}
