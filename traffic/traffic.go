package traffic

import (
	"context"
	"net/http"
	"time"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
)

// FlowControlMiddleware is HTTP middleware that wraps http.ResponseWriter
// with FlowControlledResponseWriter.
func FlowControlMiddleware(bytesPerSec int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := FlowControlResponseWriter(w, r.Context(), bytesPerSec)
			next.ServeHTTP(ww, r)
		})
	}
}

type FlowControlledResponseWriter struct {
	http.ResponseWriter
	ctx         context.Context
	bytesPerSec int
}

// FlowControlResponseWriter returns a writer that
// implements http.ResponseWriter which forces a specific
// write speed.
func FlowControlResponseWriter(w http.ResponseWriter, ctx context.Context, bytesPerSec int) *FlowControlledResponseWriter {
	return &FlowControlledResponseWriter{
		ResponseWriter: w,
		ctx:            ctx,
		bytesPerSec:    bytesPerSec,
	}
}

func (s *FlowControlledResponseWriter) Write(p []byte) (int, error) {
	// Check if ctx is already cancelled
	select {
	case <-s.ctx.Done():
		return 0, s.ctx.Err()
	default:
	}
	// Copy data in chunks
	n := 0
	max := len(p)
	for i := 0; i < max; i += s.bytesPerSec {
		j := i + s.bytesPerSec
		if j > max {
			j = max
		}
		m, err := s.ResponseWriter.Write(p[i:j])
		n += m
		if err != nil {
			return n, err
		}
		if err = s.waitN(m); err != nil {
			return n, err
		}
	}
	return n, nil
}

func (s *FlowControlledResponseWriter) waitN(n int) error {
	d := int64(n) * int64(time.Second) / int64(s.bytesPerSec)
	t := time.NewTimer(time.Duration(d))
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}
