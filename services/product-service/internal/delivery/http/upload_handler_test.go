package http

import (
	"bytes"
	"io"
	"mime/multipart"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadHandler_Success(t *testing.T) {
	uploadDir := t.TempDir()
	router := setupTestRouterFull(routerTestDeps{uploadDir: uploadDir})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "photo.png")
	assert.NoError(t, err)
	_, err = io.WriteString(part, "fake-png-content")
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	req := httptest.NewRequest(stdhttp.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "/uploads/")
}

func TestUploadHandler_MissingFile(t *testing.T) {
	router := setupTestRouterFull(routerTestDeps{uploadDir: t.TempDir()})

	req := httptest.NewRequest(stdhttp.MethodPost, "/api/upload", nil)
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusBadRequest, rec.Code)
}

func TestUploadHandler_InvalidExtension(t *testing.T) {
	router := setupTestRouterFull(routerTestDeps{uploadDir: t.TempDir()})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "doc.pdf")
	assert.NoError(t, err)
	_, err = io.WriteString(part, "pdf")
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	req := httptest.NewRequest(stdhttp.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusBadRequest, rec.Code)
}
