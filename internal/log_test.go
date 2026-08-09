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
			handler := NewEnvFilterHandler(
				slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
				targetEnvKey,
				targetEnvVal,
			)
			logger := slog.New(handler)

			test.logFunc(logger)

			got := buf.String()
			expLogSomething := len(test.expOutput) > 0
			if loggedSomething := len(got) > 0; loggedSomething != expLogSomething {
				t.Errorf("got logged=%t, expected logged=%t", loggedSomething, expLogSomething)
			}
			if expLogSomething && !strings.Contains(got, test.expOutput) {
				t.Errorf("expected output (%q) to contain %q", got, test.expOutput)
			}
		})
	}
}

func TestEnvFilterHandlerEnabled(t *testing.T) {
	// Only log when this environment variable is set to this value.
	const targetEnvKey = "TEST_DEBUG_ENV"
	const targetEnvVal = "1"

	tests := []struct {
		name     string
		envValue string
		level    slog.Level
		expected bool
	}{
		{
			name:     "Debug level disabled when env missing",
			envValue: "",
			level:    slog.LevelDebug,
			expected: false,
		},
		{
			name:     "Info level disabled when env does not match",
			envValue: "0",
			level:    slog.LevelInfo,
			expected: false,
		},
		{
			name:     "Warn level disabled when env missing",
			envValue: "",
			level:    slog.LevelWarn,
			expected: false,
		},
		{
			name:     "Error level disabled when env does not match",
			envValue: "false",
			level:    slog.LevelError,
			expected: false,
		},
		{
			name:     "Debug level enabled when env matches",
			envValue: "1",
			level:    slog.LevelDebug,
			expected: true,
		},
		{
			name:     "Info level enabled when env matches",
			envValue: "1",
			level:    slog.LevelInfo,
			expected: true,
		},
		{
			name:     "Warn level enabled when env matches",
			envValue: "1",
			level:    slog.LevelWarn,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(targetEnvKey, test.envValue)

			handler := NewEnvFilterHandler(
				slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelDebug}),
				targetEnvKey,
				targetEnvVal,
			)

			got := handler.Enabled(context.Background(), test.level)
			if got != test.expected {
				t.Errorf("handler.Enabled(%v) = %t; want %t", test.level.String(), got, test.expected)
			}
		})
	}
}
