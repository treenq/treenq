package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/treenq/treenq/pkg/vel"
)

var (
	ErrorNoSignature         = errors.New("signature is empty")
	ErrorSignaturesDontMatch = errors.New("request signatures didn't match")

	ErrNoSignatureHeader = &vel.Error{
		Code:    "NO_SIGNATURE_HEADER",
		Message: "",
	}
	NewErrInvalidSignature = func(err error) *vel.Error {
		return &vel.Error{
			Code:    "INVALID_SIGNATURE",
			Message: err.Error(),
		}
	}
)

// Sha256SignatureVerifier checks if the given payload has a valid SHA256 signature.
// It returns an error if verification fails.
type Sha256SignatureVerifier struct {
	secret          string
	signaturePrefix string
}

func NewSha256SignatureVerifier(secret string, signaturePrefix string) *Sha256SignatureVerifier {
	return &Sha256SignatureVerifier{secret: secret, signaturePrefix: signaturePrefix}
}

func (v *Sha256SignatureVerifier) Verify(payload []byte, signature string) error {
	// Ensure the signature header is provided
	if signature == "" {
		return ErrorNoSignature
	}

	// Create HMAC SHA256 hash using the secret token
	h := hmac.New(sha256.New, []byte(v.secret))
	h.Write(payload)
	expectedSignature := v.signaturePrefix + hex.EncodeToString(h.Sum(nil))

	// Compare the calculated signatures
	if !hmac.Equal([]byte(expectedSignature), []byte(signature)) {
		return ErrorSignaturesDontMatch
	}

	return nil
}

func NewSha256SignatureVerifierMiddleware(verifier *Sha256SignatureVerifier, l *slog.Logger) func(http.Handler) http.Handler {
	headerKey := "X-Hub-Signature-256"
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			signatureHeader := r.Header.Get(headerKey)
			if signatureHeader == "" {
				w.WriteHeader(http.StatusBadRequest)
				if encodeErr := json.NewEncoder(w).Encode(ErrNoSignatureHeader); encodeErr != nil {
					l.ErrorContext(r.Context(), "failed to encode error", "err", encodeErr)
				}
				return
			}

			// Read the request body to validate the signature
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "unable to read request body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// Verify the signature
			err = verifier.Verify(body, signatureHeader)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				if encodeErr := json.NewEncoder(w).Encode(NewErrInvalidSignature(err)); encodeErr != nil {
					l.ErrorContext(r.Context(), "failed to encode error", "err", encodeErr)
				}
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}
