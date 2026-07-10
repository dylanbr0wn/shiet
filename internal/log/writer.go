package log

import "io"

// RedactingWriter wraps w and scrubs sensitive JSON field values on each Write
// using the same ADR-0001 rules as RedactAttrs / ShouldRedact.
func RedactingWriter(w io.Writer) io.Writer {
	if w == nil {
		w = io.Discard
	}
	return &redactingWriter{w: w}
}

type redactingWriter struct {
	w io.Writer
}

func (rw *redactingWriter) Write(p []byte) (int, error) {
	redacted := RedactJSONLine(p)
	if _, err := rw.w.Write(redacted); err != nil {
		return 0, err
	}
	// Report the original length so callers treat the write as complete even
	// when redaction changes the byte count.
	return len(p), nil
}
