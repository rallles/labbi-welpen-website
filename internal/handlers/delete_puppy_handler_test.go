package handlers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemovePuppyUploadFilesOnlyRemovesSafeUploadTargets(t *testing.T) {
	uploadDir := t.TempDir()
	outsideDir := t.TempDir()

	insideImage := filepath.Join(uploadDir, "puppy.jpg")
	outsideImage := filepath.Join(outsideDir, "outside.jpg")
	for _, path := range []string{insideImage, outsideImage} {
		if err := os.WriteFile(path, []byte("image"), 0600); err != nil {
			t.Fatalf("os.WriteFile(%q) error = %v", path, err)
		}
	}

	paths := []string{
		"/uploads/puppy.jpg",
		"/static/outside.jpg",
		"/uploads/../../outside.jpg",
		"/uploads/missing.jpg",
	}
	if err := removePuppyUploadFiles(uploadDir, paths); err != nil {
		t.Fatalf("removePuppyUploadFiles() error = %v", err)
	}

	if _, err := os.Stat(insideImage); !os.IsNotExist(err) {
		t.Errorf("inside upload still exists or stat failed unexpectedly: %v", err)
	}
	if _, err := os.Stat(outsideImage); err != nil {
		t.Errorf("file outside upload directory was affected: %v", err)
	}
}

func TestRemovePuppyUploadFilesReportsCleanupFailure(t *testing.T) {
	uploadDir := t.TempDir()
	nonEmptyDir := filepath.Join(uploadDir, "not-a-file")
	if err := os.Mkdir(nonEmptyDir, 0700); err != nil {
		t.Fatalf("os.Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "child"), []byte("x"), 0600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if err := removePuppyUploadFiles(uploadDir, []string{"/uploads/not-a-file"}); err == nil {
		t.Fatal("removePuppyUploadFiles() error = nil, want cleanup error")
	}
}
