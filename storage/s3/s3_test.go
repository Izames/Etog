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

func createTestImage(t *testing.T) *multipart.FileHeader {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-*.jpg")
	if err != nil {
		t.Fatal(err)
	}

	content := []byte("fake image content")
	_, err = tmpFile.Write(content)
	if err != nil {
		t.Fatal(err)
	}

	err = tmpFile.Close()
	if err != nil {
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

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(10 << 20)
	if err != nil {
		t.Fatal(err)
	}

	files := req.MultipartForm.File["file"]
	if len(files) == 0 {
		t.Fatal("file not found in multipart form")
	}

	return files[0]
}

func TestUpload(t *testing.T) {

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

	fileHeader := createTestImage(t)

	url, err := s3.Upload(fileHeader, pathBucket)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	if url == "" {
		t.Fatal("expected non-empty url")
	}
}

func TestDelete(t *testing.T) {

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

	fileHeader := createTestImage(t)

	url, err := s3.Upload(fileHeader, pathBucket)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	err = s3.Delete(url)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}
