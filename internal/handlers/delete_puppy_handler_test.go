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
	nestedNameInUploadRoot := filepath.Join(uploadDir, "nested.jpg")
	traversalNameInUploadRoot := filepath.Join(uploadDir, "outside.jpg")
	outsideImage := filepath.Join(outsideDir, "outside.jpg")
	for _, path := range []string{insideImage, nestedNameInUploadRoot, traversalNameInUploadRoot, outsideImage} {
		if err := os.WriteFile(path, []byte("image"), 0600); err != nil {
			t.Fatalf("os.WriteFile(%q) error = %v", path, err)
		}
	}

	paths := []string{
		"/uploads/puppy.jpg",
		"/uploads/subdir/nested.jpg",
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
	if _, err := os.Stat(nestedNameInUploadRoot); err != nil {
		t.Errorf("nested upload path unexpectedly selected a root file: %v", err)
	}
	if _, err := os.Stat(traversalNameInUploadRoot); err != nil {
		t.Errorf("traversal upload path unexpectedly selected a root file: %v", err)
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
