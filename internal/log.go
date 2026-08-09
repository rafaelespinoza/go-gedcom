// Package internal has common code to support this module but not meant for
// use outside of the module.
package internal

import (
	"context"
	"log/slog"
	"os"
)

// EnvFilterHandler wraps an existing slog.Handler and does not log unless a
// specific environment variable is set to a preset value. When the
// environment variable does match a specific value, then the record is handed
// off to the underlying log Handler, which would determine whether or not
// it is enabled.
type EnvFilterHandler struct {
	slog.Handler
	envKey   string
	envValue string
}

// NewEnvFilterHandler creates a new filtering handler.
func NewEnvFilterHandler(handler slog.Handler, envKey, envValue string) *EnvFilterHandler {
	return &EnvFilterHandler{
		Handler:  handler,
		envKey:   envKey,
		envValue: envValue,
	}
}

// Enabled intercepts the log check. If the designates env var is equal to
// the designated value, then it lets the underlying logger decide if it's
// enabled. If the env var does not match, then it returns false.
func (h *EnvFilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if os.Getenv(h.envKey) != h.envValue {
		return false
	}

	return h.Handler.Enabled(ctx, level)
}

// Handle passes the record through if Enabled returned true.
func (h *EnvFilterHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.Handler.Handle(ctx, r)
}

// WithAttrs chains attrs when the logger calls the [*slog.Logger.With] method.
func (h *EnvFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &EnvFilterHandler{
		Handler:  h.Handler.WithAttrs(attrs),
		envKey:   h.envKey,
		envValue: h.envValue,
	}
}

// WithGroup chains a group with name, to the underlying handler's existing
// groups when the logger calls the [*slog.Logger.WithGroup] method.
func (h *EnvFilterHandler) WithGroup(name string) slog.Handler {
	return &EnvFilterHandler{
		Handler:  h.Handler.WithGroup(name),
		envKey:   h.envKey,
		envValue: h.envValue,
	}
}
