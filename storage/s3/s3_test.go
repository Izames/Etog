package s3

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"Etog/internal/config"
)

const (
	access     = "8132UvPSva3EvJ5VVB73yQ"
	secret     = "7SuUACvoJJHUHqbyhrQdjPhMXYrg5zccJ1Xe9dyxY3fL"
	bucket     = "test.bucket"
	pathBucket = "test/"
	region     = "ru-msk"
	endpoint   = "https://hb.vkcs.cloud"
)

func setupS3(t *testing.T) *S3 {
	t.Helper()
	conf := &config.S3{
		Region:   region,
		Endpoint: endpoint,
		Access:   access,
		Secret:   secret,
		Bucket:   bucket,
	}
	s3, err := NewS3(conf)
	if err != nil {
		t.Fatal(err)
	}
	return s3
}

func createTestImage(t *testing.T) *multipart.FileHeader {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	_, err = tmpFile.Write([]byte("fake image content"))
	if err != nil {
		t.Fatal(err)
	}
	if err = tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(tmpFile.Name()))
	if err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if _, err = io.Copy(part, file); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err = req.ParseMultipartForm(10 << 20); err != nil {
		t.Fatal(err)
	}

	files := req.MultipartForm.File["file"]
	if len(files) == 0 {
		t.Fatal("file not found in multipart form")
	}
	return files[0]
}

func TestUpload(t *testing.T) {
	storage := setupS3(t)
	fileHeader := createTestImage(t)

	location, err := storage.Upload(fileHeader, pathBucket)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if location == "" {
		t.Fatal("expected non-empty url")
	}

	// Проверяем что файл реально доступен
	resp, err := http.Get(location)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Errorf("файл недоступен по URL после загрузки: %v", err)
	}

	t.Cleanup(func() {
		storage.Delete(location)
	})
	t.Log("func [Upload] is OK")
}

func TestDelete(t *testing.T) {
	storage := setupS3(t)
	fileHeader := createTestImage(t)

	location, err := storage.Upload(fileHeader, pathBucket)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	err = storage.Delete(location)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Проверяем что файл реально удалён
	resp, err := http.Get(location)
	if err == nil && resp.StatusCode == http.StatusOK {
		t.Error("файл должен быть удалён, но всё ещё доступен")
		return
	}
	t.Log("func [Delete] is OK")
}
