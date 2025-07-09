package treenq

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/treenq/treenq/pkg/vel"
)

func init() {
	vel.GlobalOpts.ProcessErr = func(r *http.Request, e *vel.Error) {
		*r = *r.WithContext(LogAttrsToContext(r.Context(), "err", e.Err, "code", e.Code, "errMsg", e.Message))
	}
}

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

func (w *responseWriter) Flush() {
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}

	flusher.Flush()
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

type logAttrContextKeyType struct{}

var logAttrContextKey = logAttrContextKeyType{}

func LogAttrsToContext(ctx context.Context, kvs ...any) context.Context {
	attrsVal := ctx.Value(logAttrContextKey)

	var attrs []any
	if attrsVal == nil {
		attrs = make([]any, 0, 2)
	} else {
		attrs = attrsVal.([]any)
	}
	attrs = append(attrs, kvs...)
	return context.WithValue(ctx, logAttrContextKey, attrs)
}

func LogAtrtsFromContext(ctx context.Context) []any {
	attrsVal := ctx.Value(logAttrContextKey)
	if attrsVal == nil {
		return nil
	}
	return attrsVal.([]any)
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
			resp := &responseWriter{ResponseWriter: w}

			defer func() {
				if recovered := recover(); recovered != nil {
					l.ErrorContext(
						r.Context(), "recovered from panic",
						"uri", r.RequestURI,
						"method", r.Method,
						"recovered", recovered,
						"stack", string(debug.Stack()),
						"time", time.Now().Format(time.RFC3339),
					)
					resp.WriteHeader(http.StatusInternalServerError)
					if _, err := resp.Write([]byte("internal server error")); err != nil {
						l.ErrorContext(r.Context(), "failed to write response", "error", err)
					}
				}
			}()

			start := time.Now()

			ctxWithLogger := LoggerToContext(r.Context(), l)
			r = r.WithContext(ctxWithLogger)
			next.ServeHTTP(resp, r)

			end := time.Now()
			duration := end.Sub(start)
			// soon we can log the url pattern and easy to match it to our observability toolings
			// https://github.com/golang/go/issues/66405
			// but now let's enjoy RequestURI
			var logFunc func(context.Context, string, ...any)
			if resp.status >= 500 {
				logFunc = l.ErrorContext
			} else if resp.status >= 400 {
				logFunc = l.InfoContext
			} else {
				logFunc = l.DebugContext
			}

			attrs := []any{
				"duration", duration.String(),
				"time", end.Format(time.RFC3339),
				"status", resp.status,
				"uri", r.RequestURI,
				"method", r.Method,
			}
			attrs = append(attrs, LogAtrtsFromContext(r.Context())...)

			logFunc(
				r.Context(), "request completed",
				attrs...,
			)
		})
	}
}
