package vel

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLog(t *testing.T) {
	type testCase struct {
		name    string
		handler http.Handler

		expectedLogFunc func(t *testing.T, m map[string]any)
		expectedStatus  int
	}

	for _, tt := range []testCase{
		{
			name: "regular log",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				LoggerFromContext(r.Context()).ErrorContext(r.Context(), "test", "arg1", "text")
				w.WriteHeader(200)
			}),
			expectedLogFunc: func(t *testing.T, m map[string]any) {
				t.Helper()

				assert.Equal(t, "test", m["msg"])
				assert.Equal(t, "text", m["arg1"])
				assert.Equal(t, "ERROR", m["level"])

				timeStr := m["time"].(string)
				timeValue, err := time.Parse(time.RFC3339, timeStr)
				require.NoError(t, err)
				// make sure it's not empty and not broken
				assert.True(t, timeValue.After(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)))
			},
			expectedStatus: 200,
		},
		{
			name: "append attrs to context",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := LogAttrsToContext(r.Context(), "key", "42")
				*r = *r.WithContext(ctx)
				// 400 must trigger info level
				w.WriteHeader(400)
			}),
			expectedLogFunc: func(t *testing.T, m map[string]any) {
				t.Helper()

				assert.Equal(t, "42", m["key"])
				assert.Equal(t, "INFO", m["level"])

				timeStr := m["time"].(string)
				timeValue, err := time.Parse(time.RFC3339, timeStr)
				require.NoError(t, err)
				// make sure it's not empty and not broken
				assert.True(t, timeValue.After(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)))
			},
			expectedStatus: 400,
		},
		{
			name: "panic log",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("test")
			}),
			expectedLogFunc: func(t *testing.T, m map[string]any) {
				t.Helper()

				assert.Equal(t, "recovered from panic", m["msg"])
				assert.Equal(t, "test", m["recovered"])
				assert.Equal(t, "ERROR", m["level"])
				assert.Equal(t, "/", m["uri"])
				assert.Equal(t, "POST", m["method"])

				timeStr := m["time"].(string)
				timeValue, err := time.Parse(time.RFC3339, timeStr)
				require.NoError(t, err)
				// make sure it's not empty and not broken
				assert.True(t, timeValue.After(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)))
				// make sure it's not empty and non zero value
				assert.True(t, strings.HasPrefix(m["stack"].(string), "goroutine"))
			},
			expectedStatus: 500,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			l := NewLogger(buf, slog.LevelInfo)
			m := NewLoggingMiddleware(l)
			h := m(tt.handler)

			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("POST", "/", nil))

			res := buf.String()
			jsonRes := make(map[string]any, 0)
			err := json.Unmarshal([]byte(res), &jsonRes)
			require.NoError(t, err)

			tt.expectedLogFunc(t, jsonRes)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
