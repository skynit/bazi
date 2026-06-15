package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newRouter() *gin.Engine {
	r := gin.New()
	r.GET("/data", ETag(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"hello": "world", "n": 42})
	})
	return r
}

func TestETag_FirstRequest_HasETag(t *testing.T) {
	r := newRouter()

	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d", w.Code)
	}
	if w.Header().Get("ETag") == "" {
		t.Fatal("ETag header missing on first response")
	}
	if w.Header().Get("Cache-Control") == "" {
		t.Fatal("Cache-Control missing")
	}
}

func TestETag_Deterministic(t *testing.T) {
	r := newRouter()

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/data", nil))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/data", nil))

	e1 := w1.Header().Get("ETag")
	e2 := w2.Header().Get("ETag")
	if e1 == "" || e1 != e2 {
		t.Fatalf("ETag mismatch across identical requests: %q vs %q", e1, e2)
	}
}

func TestETag_IfNoneMatch_Returns304(t *testing.T) {
	r := newRouter()

	first := httptest.NewRecorder()
	r.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/data", nil))
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("could not obtain ETag from first request")
	}

	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	req.Header.Set("If-None-Match", etag)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotModified {
		t.Fatalf("status: want 304, got %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("304 body must be empty, got %d bytes", w.Body.Len())
	}
}

func TestETag_NonMatchingIfNoneMatch_Returns200(t *testing.T) {
	r := newRouter()

	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	req.Header.Set("If-None-Match", `"deadbeef"`)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Fatal("body should be present on 200")
	}
}
