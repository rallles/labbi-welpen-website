package handlers

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveUploadedImagesAllowsJPEGAndPNG(t *testing.T) {
	uploadDir := t.TempDir()
	files := multipartFiles(t,
		testFile{name: "photo.jpg", data: jpegImage(t)},
		testFile{name: "photo.png", data: pngImage(t)},
	)

	paths, err := saveUploadedImages(files, uploadDir)
	if err != nil {
		t.Fatalf("saveUploadedImages() error = %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("got %d paths, want 2", len(paths))
	}
	for _, path := range paths {
		if !strings.HasPrefix(path, "/uploads/") {
			t.Fatalf("path %q does not use /uploads prefix", path)
		}
		if !fileExists(filepath.Join(uploadDir, filepath.Base(path))) {
			t.Fatalf("uploaded file for %q does not exist", path)
		}
	}
}

func TestSaveUploadedImagesRejectsNonImages(t *testing.T) {
	files := multipartFiles(t, testFile{name: "not-image.jpg", data: []byte("<svg></svg>")})

	_, err := saveUploadedImages(files, t.TempDir())
	if err == nil {
		t.Fatal("saveUploadedImages() error = nil, want error")
	}
	if uploadErrorStatus(err) != http.StatusBadRequest {
		t.Fatalf("uploadErrorStatus() = %d, want %d", uploadErrorStatus(err), http.StatusBadRequest)
	}
}

func TestSaveUploadedImagesRejectsOversizedFile(t *testing.T) {
	files := multipartFiles(t, testFile{name: "large.png", data: bytes.Repeat([]byte{0xff}, maxUploadFileSize+1)})

	_, err := saveUploadedImages(files, t.TempDir())
	if err == nil {
		t.Fatal("saveUploadedImages() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("error = %v, want size limit error", err)
	}
}

func TestSaveUploadedImagesRejectsTooManyImages(t *testing.T) {
	files := make([]testFile, maxUploadImages+1)
	for i := range files {
		files[i] = testFile{name: "photo.png", data: pngImage(t)}
	}

	_, err := saveUploadedImages(multipartFiles(t, files...), t.TempDir())
	if err == nil {
		t.Fatal("saveUploadedImages() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "too many images") {
		t.Fatalf("error = %v, want image count limit error", err)
	}
}

func TestSaveUploadedImagesUsesSafeServerFilenames(t *testing.T) {
	uploadDir := t.TempDir()
	files := multipartFiles(t, testFile{name: "../../evil.php.png", data: pngImage(t)})

	paths, err := saveUploadedImages(files, uploadDir)
	if err != nil {
		t.Fatalf("saveUploadedImages() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(paths))
	}

	path := paths[0]
	base := filepath.Base(path)
	if path != "/uploads/"+base {
		t.Fatalf("path %q is not a simple /uploads basename path", path)
	}
	if strings.Contains(base, "evil") || strings.Contains(base, "..") || strings.Contains(base, "/") || strings.Contains(base, "\\") {
		t.Fatalf("filename %q contains user-controlled path content", base)
	}
	if filepath.Ext(base) != ".png" {
		t.Fatalf("extension = %q, want .png", filepath.Ext(base))
	}
}

type testFile struct {
	name string
	data []byte
}

func multipartFiles(t *testing.T, files ...testFile) []*multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, file := range files {
		part, err := writer.CreateFormFile("images", file.name)
		if err != nil {
			t.Fatalf("CreateFormFile() error = %v", err)
		}
		if _, err := part.Write(file.data); err != nil {
			t.Fatalf("part.Write() error = %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/", &body)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(multipartMemory); err != nil {
		t.Fatalf("ParseMultipartForm() error = %v", err)
	}
	return req.MultipartForm.File["images"]
}

func pngImage(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, testImage()); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}
	return buf.Bytes()
}

func jpegImage(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, testImage(), nil); err != nil {
		t.Fatalf("jpeg.Encode() error = %v", err)
	}
	return buf.Bytes()
}

func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	return img
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
