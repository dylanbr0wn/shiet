// Package log provides a shared zerolog logger with ADR-0001 secret redaction
// for both the desktop app and the OAuth broker.
package log

import (
	"io"

	"github.com/rs/zerolog"
)

// RedactHook is a zerolog.Hook that marks a logger as using ADR-0001 redaction.
// Field scrubbing runs in RedactingWriter because zerolog encodes fields into
// the event buffer before hooks execute. Prefer New, which attaches both.
type RedactHook struct{}

// Run implements zerolog.Hook. Redaction is applied by RedactingWriter.
func (RedactHook) Run(_ *zerolog.Event, _ zerolog.Level, _ string) {}

// New returns a JSON Info-level zerolog.Logger with timestamps and secret
// redaction (sensitive keys and token-shaped values → [redacted]).
func New(w io.Writer) zerolog.Logger {
	if w == nil {
		w = io.Discard
	}
	return zerolog.New(RedactingWriter(w)).
		Level(zerolog.InfoLevel).
		With().Timestamp().
		Logger().
		Hook(RedactHook{})
}
