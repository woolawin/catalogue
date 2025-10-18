package ext

import (
	"io"
	"path/filepath"
	"slices"
	"strings"
)

type MockDisk struct {
	Dirs  []string
	Files []string
}

func (mock *MockDisk) Path(parts ...string) string {
	return filepath.Join(parts...)
}

func (mock *MockDisk) FileExists(path string) (bool, bool, error) {
	if slices.Contains(mock.Files, path) {
		return true, true, nil
	}
	return false, false, nil
}

func (mock *MockDisk) DirExists(path string) (bool, bool, error) {
	if slices.Contains(mock.Dirs, path) {
		return true, true, nil
	}
	return false, false, nil
}

func (mock *MockDisk) List(path string) ([]string, []string, error) {
	var files []string
	var dirs []string

	for _, file := range mock.Files {
		if file == path {
			continue
		}
		if strings.HasPrefix(file, path) {
			relative, _ := strings.CutPrefix(file, path+"/")
			files = append(files, relative)
		}
	}

	for _, dir := range mock.Dirs {
		if dir == path {
			continue
		}
		if strings.HasPrefix(dir, path) {
			relative, _ := strings.CutPrefix(dir, path+"/")
			dirs = append(dirs, relative)
		}
	}

	return files, dirs, nil
}

func (mock *MockDisk) ArchiveDir(src, dst string) error {
	return nil
}

func (mock *MockDisk) WriteFile(path string, data io.Reader) error {
	return nil
}

func (mock *MockDisk) CreateDir(path string) error {
	mock.Dirs = append(mock.Dirs, path)
	return nil
}

func (mock *MockDisk) ListRec(path string) ([]string, error) {
	return nil, nil
}

func (mock *MockDisk) CreateTar(path string) error {
	mock.Files = append(mock.Files, path)
	return nil
}

func (mock *MockDisk) Move(toPath string, fromPath string, files []string, overwrite bool) error {
	return nil
}

func (mock *MockDisk) ArchiveFiles(path string, files []string) error {
	return nil
}
