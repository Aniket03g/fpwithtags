package log

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogger(t *testing.T) {
	// Test that logger is initialized
	logger := Logger()
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// Test that logger has JSON formatter
	if _, ok := logger.Formatter.(*logrus.JSONFormatter); !ok {
		t.Error("Logger should use JSONFormatter")
	}

	// Test default log level
	level := GetLevel()
	if level != logrus.InfoLevel {
		t.Errorf("Expected default log level to be Info, got %s", level)
	}
}

func TestSetLevel(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	// Test setting different levels
	levels := []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
	}

	for _, level := range levels {
		SetLevel(level)
		if GetLevel() != level {
			t.Errorf("Expected level %s, got %s", level, GetLevel())
		}
	}
}

func TestLoggingFunctions(t *testing.T) {
	// These tests just ensure the functions don't panic
	// Actual output verification would require capturing stdout/file

	t.Run("Info", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Info panicked: %v", r)
			}
		}()
		Info("test info message")
		Infof("test info message with format: %s", "value")
	})

	t.Run("Debug", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Debug panicked: %v", r)
			}
		}()
		Debug("test debug message")
		Debugf("test debug message with format: %s", "value")
	})

	t.Run("Warn", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Warn panicked: %v", r)
			}
		}()
		Warn("test warn message")
		Warnf("test warn message with format: %s", "value")
	})

	t.Run("Error", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Error panicked: %v", r)
			}
		}()
		Error("test error message")
		Errorf("test error message with format: %s", "value")
	})
}

func TestWithFields(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("WithFields panicked: %v", r)
		}
	}()

	entry := WithFields(logrus.Fields{
		"key1": "value1",
		"key2": 123,
	})

	if entry == nil {
		t.Error("WithFields should return a non-nil entry")
	}
}

func TestWithField(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("WithField panicked: %v", r)
		}
	}()

	entry := WithField("key", "value")

	if entry == nil {
		t.Error("WithField should return a non-nil entry")
	}
}

func TestWithError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("WithError panicked: %v", r)
		}
	}()

	err := &testError{msg: "test error"}
	entry := WithError(err)

	if entry == nil {
		t.Error("WithError should return a non-nil entry")
	}
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
