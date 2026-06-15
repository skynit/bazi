package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

// etagWriter buffers status, headers, and body so we can compute an ETag and
// optionally short-circuit to 304 before any bytes hit the wire.
type etagWriter struct {
	gin.ResponseWriter
	buf      *bytes.Buffer
	status   int
	wroteHdr bool
}

func (w *etagWriter) WriteHeader(code int) {
	w.status = code
	w.wroteHdr = true
	// Defer writing to the wrapped writer until ETag has been decided.
}

func (w *etagWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *etagWriter) Write(p []byte) (int, error)         { return w.buf.Write(p) }
func (w *etagWriter) WriteString(s string) (int, error)   { return w.buf.WriteString(s) }
func (w *etagWriter) Written() bool                       { return w.wroteHdr || w.buf.Len() > 0 }
func (w *etagWriter) Size() int                           { return w.buf.Len() }

// ETag wraps a handler so successful (200) JSON responses gain a deterministic
// ETag header derived from the body. Requests sending a matching If-None-Match
// receive 304 with no body.
//
// Cache-Control is set to "private, max-age=300" — fortunes are user-specific
// (JWT-bound) so they must not be shared across users.
func ETag() gin.HandlerFunc {
	return func(c *gin.Context) {
		original := c.Writer
		ew := &etagWriter{ResponseWriter: original, buf: &bytes.Buffer{}}
		c.Writer = ew

		c.Next()

		// Restore original writer so subsequent middlewares (and Gin's own
		// finalize logic) operate on the real connection.
		c.Writer = original

		status := ew.Status()
		body := ew.buf.Bytes()

		// Only stamp ETag for plain 200 responses with a body.
		if status != http.StatusOK || len(body) == 0 {
			if ew.wroteHdr {
				original.WriteHeader(status)
			}
			if len(body) > 0 {
				_, _ = original.Write(body)
			}
			return
		}

		sum := sha256.Sum256(body)
		etag := `"` + hex.EncodeToString(sum[:8]) + `"`

		original.Header().Set("ETag", etag)
		original.Header().Set("Cache-Control", "private, max-age=300")

		if match := c.GetHeader("If-None-Match"); match == etag {
			original.WriteHeader(http.StatusNotModified)
			return
		}

		original.WriteHeader(http.StatusOK)
		_, _ = original.Write(body)
	}
}
