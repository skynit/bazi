package handler

import (
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"path/filepath"
	"strings"
	"testing"

	"bazi/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type elementAssetTestWriter struct {
	asset *model.ElementAsset
	err   error
}

func (w *elementAssetTestWriter) Create(asset *model.ElementAsset) error {
	w.asset = asset
	return w.err
}

type elementAssetTestUsers struct {
	user *model.User
	err  error
}

func (s elementAssetTestUsers) FindByID(uint) (*model.User, error) {
	return s.user, s.err
}

func TestElementAssetUploadRequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ElementAssetHandler{
		Writer:    &elementAssetTestWriter{},
		Users:     elementAssetTestUsers{user: &model.User{Username: "visitor"}},
		UploadDir: t.TempDir(),
	}

	response := performElementAssetUpload(h, 1, nil, "")
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestElementAssetUploadRequiresFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newElementAssetUploadTestHandler(t, &elementAssetTestWriter{})
	body, contentType := elementAssetMultipart(t, map[string]string{
		"element":  "木",
		"name":     "晨林",
		"alt_text": "晨光中的树林",
	}, nil, "")

	response := performElementAssetUpload(h, 1, body, contentType)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestElementAssetUploadDetectsOrientation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name        string
		width       int
		height      int
		orientation string
	}{
		{name: "square", width: 4, height: 4, orientation: "square"},
		{name: "landscape", width: 8, height: 4, orientation: "landscape"},
		{name: "portrait", width: 4, height: 8, orientation: "portrait"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writer := &elementAssetTestWriter{}
			h := newElementAssetUploadTestHandler(t, writer)
			imageBytes := encodeElementAssetPNG(t, tc.width, tc.height)
			body, contentType := elementAssetMultipart(t, map[string]string{
				"element":  "木",
				"name":     "晨林",
				"alt_text": "晨光中的树林",
			}, imageBytes, "image/png")

			response := performElementAssetUpload(h, 1, body, contentType)
			if response.Code != http.StatusCreated {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
			}
			if writer.asset == nil {
				t.Fatal("asset metadata was not saved")
			}
			if writer.asset.Orientation != tc.orientation {
				t.Fatalf("orientation = %q, want %q", writer.asset.Orientation, tc.orientation)
			}
			if writer.asset.Width != tc.width || writer.asset.Height != tc.height {
				t.Fatalf("dimensions = %dx%d, want %dx%d", writer.asset.Width, writer.asset.Height, tc.width, tc.height)
			}
			if !strings.HasPrefix(writer.asset.URL, "/uploads/element-assets/wood/") {
				t.Fatalf("unexpected asset URL %q", writer.asset.URL)
			}
			matches, err := filepath.Glob(filepath.Join(h.UploadDir, "wood", "*.png"))
			if err != nil || len(matches) != 1 {
				t.Fatalf("uploaded files = %v, err=%v", matches, err)
			}
		})
	}
}

func TestElementAssetUploadRejectsInvalidImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	writer := &elementAssetTestWriter{}
	h := newElementAssetUploadTestHandler(t, writer)
	body, contentType := elementAssetMultipart(t, map[string]string{
		"element":  "火",
		"name":     "伪图片",
		"alt_text": "无效图片",
	}, []byte("not an image"), "image/png")

	response := performElementAssetUpload(h, 1, body, contentType)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if writer.asset != nil {
		t.Fatal("invalid image metadata should not be saved")
	}
}

func TestElementAssetUploadRejectsOversizedFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	writer := &elementAssetTestWriter{}
	h := newElementAssetUploadTestHandler(t, writer)
	body, contentType := elementAssetMultipart(t, map[string]string{
		"element":  "水",
		"name":     "过大图片",
		"alt_text": "过大的图片文件",
	}, make([]byte, maxElementAssetBytes+1), "image/png")

	response := performElementAssetUpload(h, 1, body, contentType)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if writer.asset != nil {
		t.Fatal("oversized image metadata should not be saved")
	}
}

func newElementAssetUploadTestHandler(t *testing.T, writer *elementAssetTestWriter) *ElementAssetHandler {
	t.Helper()
	return &ElementAssetHandler{
		Writer: writer,
		Users: elementAssetTestUsers{user: &model.User{
			Model:    gorm.Model{ID: 1},
			Username: "admin",
		}},
		AdminUsername: "admin",
		UploadDir:     t.TempDir(),
	}
}

func performElementAssetUpload(h *ElementAssetHandler, userID uint, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	if body == nil {
		body = &bytes.Buffer{}
	}
	request := httptest.NewRequest(http.MethodPost, "/api/element-assets", body)
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = request
	context.Set("userID", userID)
	h.Upload(context)
	return response
}

func elementAssetMultipart(t *testing.T, fields map[string]string, file []byte, fileContentType string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if file != nil {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="file"; filename="asset.png"`)
		header.Set("Content-Type", fileContentType)
		part, err := writer.CreatePart(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(file); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}

func encodeElementAssetPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
