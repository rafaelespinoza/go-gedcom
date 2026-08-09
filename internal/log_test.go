package internal

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestEnvFilterHandler(t *testing.T) {
	// Only log when this environment variable is set to this value.
	const targetEnvKey = "TEST_DEBUG_APP"
	const targetEnvVal = "true"

	tests := []struct {
		name      string
		envValue  string
		logFunc   func(l *slog.Logger)
		expOutput string
	}{
		{
			name:     "INFO log suppressed when env var is unset",
			envValue: "",
			logFunc: func(l *slog.Logger) {
				l.Info("info message")
			},
		},
		{
			name:     "WARN log suppressed when env var is wrong value",
			envValue: "false",
			logFunc: func(l *slog.Logger) {
				l.Warn("warn message")
			},
		},
		{
			name:     "ERROR log suppressed when env var is wrong value",
			envValue: "off",
			logFunc: func(l *slog.Logger) {
				l.Error("error message")
			},
		},
		{
			name:     "DEBUG log suppressed when env var is unset",
			envValue: "",
			logFunc: func(l *slog.Logger) {
				l.Debug("debug message")
			},
		},
		{
			name:     "INFO log emitted when env var matches target value",
			envValue: "true",
			logFunc: func(l *slog.Logger) {
				l.Info("info message")
			},
			expOutput: "info message",
		},
		{
			name:     "DEBUG log emitted when env var matches target value",
			envValue: "true",
			logFunc: func(l *slog.Logger) {
				l.Debug("debug message")
			},
			expOutput: "debug message",
		},
		{
			name:     "ERROR log emitted when env var matches target value",
			envValue: "true",
			logFunc: func(l *slog.Logger) {
				l.Error("critical error")
			},
			expOutput: "critical error",
		},
		{
			name:     "Logger attributes preserved through WithAttrs when enabled",
			envValue: "true",
			logFunc: func(l *slog.Logger) {
				l.With("user_id", 42).Debug("user debug message")
			},
			expOutput: "user_id=42",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(targetEnvKey, test.envValue)

			var buf bytes.Buffer
			baseHandler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})
			handler := NewEnvFilterHandler(baseHandler, targetEnvKey, targetEnvVal)
			logger := slog.New(handler)

			test.logFunc(logger)

			gotOutput := buf.String()

			expLogSomething := len(test.expOutput) > 0
			if loggedSomething := len(gotOutput) > 0; loggedSomething != expLogSomething {
				t.Errorf("got logged=%t, expected logged=%t", loggedSomething, expLogSomething)
			}

			if expLogSomething && !strings.Contains(gotOutput, test.expOutput) {
				t.Errorf("expected output (%q) to contain %q", gotOutput, test.expOutput)
			}
		})
	}
}

func TestEnvFilterHandlerEnabled(t *testing.T) {
	// Only log when this environment variable is set to this value.
	const targetEnvKey = "TEST_DEBUG_ENV"
	const targetEnvVal = "1"

	tests := []struct {
		name          string
		envValue      string
		checkLevel    slog.Level
		expectedState bool
	}{
		{
			name:          "Info level disabled when env false",
			envValue:      "0",
			checkLevel:    slog.LevelInfo,
			expectedState: false,
		},
		{
			name:          "Debug level disabled when env missing",
			envValue:      "",
			checkLevel:    slog.LevelDebug,
			expectedState: false,
		},
		{
			name:          "Warn level disabled when env missing",
			envValue:      "",
			checkLevel:    slog.LevelWarn,
			expectedState: false,
		},
		{
			name:          "Error level disabled when env wrong",
			envValue:      "false",
			checkLevel:    slog.LevelError,
			expectedState: false,
		},
		{
			name:          "Debug level enabled when env matches",
			envValue:      "1",
			checkLevel:    slog.LevelDebug,
			expectedState: true,
		},
		{
			name:          "Info level enabled when env matches",
			envValue:      "1",
			checkLevel:    slog.LevelInfo,
			expectedState: true,
		},
		{
			name:          "Warn level enabled when env matches",
			envValue:      "1",
			checkLevel:    slog.LevelWarn,
			expectedState: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(targetEnvKey, test.envValue)

			baseHandler := slog.NewTextHandler(nil, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})
			handler := NewEnvFilterHandler(baseHandler, targetEnvKey, targetEnvVal)

			enabled := handler.Enabled(context.Background(), test.checkLevel)
			if enabled != test.expectedState {
				t.Errorf("handler.Enabled(%v) = %t; want %t", test.checkLevel.String(), enabled, test.expectedState)
			}
		})
	}
}
