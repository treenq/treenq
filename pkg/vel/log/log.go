package log

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

type responseWriter struct {
	http.ResponseWriter

	status      int
	wroteHeader bool
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.status = 200
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(p)
	return n, err
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if !w.wroteHeader {
		w.wroteHeader = true
		w.status = statusCode
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

type logContextKey struct{}

var DefaultLogWriter = os.Stderr

func LoggerFromContext(ctx context.Context) *slog.Logger {
	v := ctx.Value(logContextKey{})
	l, ok := v.(*slog.Logger)
	if !ok {
		l := NewLogger(DefaultLogWriter, slog.LevelInfo)
		l.Info("no logger found in context")
	}
	return l
}

func LoggerToContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, logContextKey{}, l)
}

func Formatter(groups []string, a slog.Attr) slog.Attr {
	if a.Key != slog.TimeKey {
		return a
	}

	if a.Value.Kind() == slog.KindTime {
		t := a.Value.Time()
		a.Value = slog.StringValue(t.Format(time.RFC3339))
	}
	return a
}

func NewLogger(w io.Writer, level slog.Level) *slog.Logger {
	logHandler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level, ReplaceAttr: Formatter})
	return slog.New(logHandler)
}

func NewLoggingMiddleware(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// all the magic of opentelemetry could be here
			logger := l.With("service", "userService")

			resp := &responseWriter{ResponseWriter: w}

			defer func() {
				if recovered := recover(); recovered != nil {
					logger.ErrorContext(
						r.Context(), "recovered from panic",
						"uri", r.RequestURI,
						"method", r.Method,
						"recovered", recovered,
						"stack", string(debug.Stack()),
					)
					resp.WriteHeader(http.StatusInternalServerError)
					if _, err := resp.Write([]byte("internal server error")); err != nil {
						logger.ErrorContext(r.Context(), "failed to write response", "error", err)
					}
				}
			}()

			start := time.Now()
			logger.DebugContext(
				r.Context(), "request received",
				"uri", r.RequestURI,
				"method", r.Method,
			)

			ctxWithLogger := LoggerToContext(r.Context(), logger)
			r = r.WithContext(ctxWithLogger)
			next.ServeHTTP(resp, r)

			end := time.Now()
			duration := end.Sub(start)
			// soon we can log the url pattern and easy to match it to our observability toolings
			// https://github.com/golang/go/issues/66405
			// but now let's enjoy RequestURI
			var logFunc func(context.Context, string, ...any)
			if resp.status >= 500 {
				logFunc = logger.ErrorContext
			} else if resp.status >= 400 {
				logFunc = logger.InfoContext
			} else {
				logFunc = logger.DebugContext
			}

			logFunc(
				r.Context(), "request completed",
				"duration", duration.String(),
				"time", end.Format(time.RFC3339),
				"status", resp.status,
			)
		})
	}
}
