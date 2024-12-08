package vel

import (
	"context"
	"net/http"
)

type (
	requestKeyType string
	writerKeyType  string
)

const (
	requestKey requestKeyType = "requestKey"
	writerKey  writerKeyType  = "writerKey"
)

func RequestWithContext(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestKey, r)
}

func RequestFromContext(ctx context.Context) *http.Request {
	if r, ok := ctx.Value(requestKey).(*http.Request); ok {
		return r
	}
	return nil
}

func WriterWithContext(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, writerKey, w)
}

func WriterFromContext(ctx context.Context) http.ResponseWriter {
	if w, ok := ctx.Value(writerKey).(http.ResponseWriter); ok {
		return w
	}
	return nil
}
