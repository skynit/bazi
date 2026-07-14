package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bazi/internal/model"
	"bazi/internal/service/elementasset"

	"github.com/gin-gonic/gin"
)

const maxElementAssetBytes = 12 << 20

type ElementAssetSelector interface {
	Select(query elementasset.Query) ([]model.ElementAsset, error)
}

type ElementAssetWriter interface {
	Create(asset *model.ElementAsset) error
}

type ElementAssetUserStore interface {
	FindByID(id uint) (*model.User, error)
}

type ElementAssetHandler struct {
	Assets        ElementAssetSelector
	Writer        ElementAssetWriter
	Users         ElementAssetUserStore
	AdminUsername string
	UploadDir     string
}

// Select handles GET /api/element-assets/select.
func (h *ElementAssetHandler) Select(c *gin.Context) {
	if h == nil || h.Assets == nil {
		respondError(c, http.StatusServiceUnavailable, ErrCodeServiceDisabled, "element asset library is not available")
		return
	}

	limit := 8
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 30 {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "limit must be between 1 and 30")
			return
		}
		limit = parsed
	}

	assets, err := h.Assets.Select(elementasset.Query{
		Element:     strings.TrimSpace(c.Query("element")),
		Scene:       strings.TrimSpace(c.Query("scene")),
		Orientation: strings.TrimSpace(c.Query("orientation")),
		Season:      strings.TrimSpace(c.Query("season")),
		TimePeriod:  strings.TrimSpace(c.Query("time_period")),
		Seed:        strings.TrimSpace(c.Query("seed")),
		Limit:       limit,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to select element assets")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"assets": assets})
}

// Upload handles POST /api/element-assets. It is intentionally restricted to
// the configured admin account because the project does not yet have roles.
func (h *ElementAssetHandler) Upload(c *gin.Context) {
	if h == nil || h.Writer == nil || h.Users == nil || strings.TrimSpace(h.UploadDir) == "" {
		respondError(c, http.StatusServiceUnavailable, ErrCodeServiceDisabled, "element asset upload is not available")
		return
	}
	uid, ok := authUserID(c)
	if !ok || !h.isAdmin(uid) {
		respondError(c, http.StatusForbidden, ErrCodeUnauthorized, "admin access is required")
		return
	}

	element := strings.TrimSpace(c.PostForm("element"))
	if !validAssetElement(element) {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "element must be one of 木、火、土、金、水")
		return
	}
	name := strings.TrimSpace(c.PostForm("name"))
	altText := strings.TrimSpace(c.PostForm("alt_text"))
	if name == "" || altText == "" {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "name and alt_text are required")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "image file is required")
		return
	}
	defer file.Close()
	if header.Size <= 0 || header.Size > maxElementAssetBytes {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "image must be smaller than 12MB")
		return
	}

	ext, ok := assetExtension(header)
	if !ok {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "only PNG, JPEG and GIF images are supported")
		return
	}
	key, err := randomAssetKey(element)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to generate asset key")
		return
	}
	directory := filepath.Join(h.UploadDir, assetElementDirectory(element))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to prepare asset directory")
		return
	}
	filename := key + ext
	path := filepath.Join(directory, filename)
	width, height, err := saveAndInspectImage(file, path)
	if err != nil {
		_ = os.Remove(path)
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid image file")
		return
	}

	orientation := strings.TrimSpace(c.PostForm("orientation"))
	if orientation == "" {
		orientation = assetOrientation(width, height)
	}
	asset := &model.ElementAsset{
		Key:              key,
		Name:             name,
		Element:          element,
		SecondaryElement: strings.TrimSpace(c.PostForm("secondary_element")),
		URL:              fmt.Sprintf("/uploads/element-assets/%s/%s", assetElementDirectory(element), filename),
		Scene:            formDefault(c, "scene", "general"),
		Orientation:      orientation,
		Tone:             formDefault(c, "tone", "balanced"),
		Style:            formDefault(c, "style", "chinese"),
		Season:           formDefault(c, "season", "all"),
		TimePeriod:       formDefault(c, "time_period", "all"),
		ObjectLabel:      strings.TrimSpace(c.PostForm("object_label")),
		Description:      strings.TrimSpace(c.PostForm("description")),
		AltText:          altText,
		DominantColor:    strings.TrimSpace(c.PostForm("dominant_color")),
		AccentColor:      strings.TrimSpace(c.PostForm("accent_color")),
		FocalX:           formFloat(c, "focal_x", 0.5),
		FocalY:           formFloat(c, "focal_y", 0.5),
		Width:            width,
		Height:           height,
		Weight:           formInt(c, "weight", 100),
		Status:           model.ElementAssetStatusActive,
	}
	if err := h.Writer.Create(asset); err != nil {
		_ = os.Remove(path)
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to save asset metadata")
		return
	}
	respondJSON(c, http.StatusCreated, asset)
}

func (h *ElementAssetHandler) isAdmin(userID uint) bool {
	user, err := h.Users.FindByID(userID)
	if err != nil || user == nil {
		return false
	}
	admin := strings.TrimSpace(h.AdminUsername)
	if admin == "" {
		admin = "admin"
	}
	return user.Username == admin
}

func RegisterElementAssetRoutes(r gin.IRouter, assets ElementAssetSelector, writer ElementAssetWriter, users ElementAssetUserStore, adminUsername, uploadDir string) {
	h := &ElementAssetHandler{Assets: assets, Writer: writer, Users: users, AdminUsername: adminUsername, UploadDir: uploadDir}
	r.GET("/element-assets/select", h.Select)
	r.POST("/element-assets", h.Upload)
}

func validAssetElement(element string) bool {
	switch element {
	case "木", "火", "土", "金", "水":
		return true
	default:
		return false
	}
}

func assetExtension(header *multipart.FileHeader) (string, bool) {
	switch strings.ToLower(header.Header.Get("Content-Type")) {
	case "image/png":
		return ".png", true
	case "image/jpeg", "image/jpg":
		return ".jpg", true
	case "image/gif":
		return ".gif", true
	default:
		return "", false
	}
}

func randomAssetKey(element string) (string, error) {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return assetElementDirectory(element) + "-" + hex.EncodeToString(random[:]), nil
}

func assetElementDirectory(element string) string {
	return map[string]string{"木": "wood", "火": "fire", "土": "earth", "金": "metal", "水": "water"}[element]
}

func saveAndInspectImage(source multipart.File, target string) (int, int, error) {
	out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, 0, err
	}
	written, copyErr := io.Copy(out, io.LimitReader(source, maxElementAssetBytes+1))
	if copyErr != nil {
		_ = out.Close()
		return 0, 0, copyErr
	}
	if written > maxElementAssetBytes {
		_ = out.Close()
		return 0, 0, fmt.Errorf("image exceeds %d bytes", maxElementAssetBytes)
	}
	if err = out.Close(); err != nil {
		return 0, 0, err
	}
	file, err := os.Open(target)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}
	return config.Width, config.Height, nil
}

func assetOrientation(width, height int) string {
	if width == height {
		return "square"
	}
	if width > height*2 {
		return "panorama"
	}
	if width > height {
		return "landscape"
	}
	return "portrait"
}

func formDefault(c *gin.Context, key, fallback string) string {
	if value := strings.TrimSpace(c.PostForm(key)); value != "" {
		return value
	}
	return fallback
}

func formInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(c.PostForm(key)))
	if err != nil {
		return fallback
	}
	return value
}

func formFloat(c *gin.Context, key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(c.PostForm(key)), 64)
	if err != nil || value < 0 || value > 1 {
		return fallback
	}
	return value
}
