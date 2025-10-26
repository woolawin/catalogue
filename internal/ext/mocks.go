package ext

import (
	"io"
	"path/filepath"
	"slices"
	"strings"

	"github.com/woolawin/catalogue/internal"
)

type MockDisk struct {
	Dirs  []string
	Files []string
}

func (mock *MockDisk) Path(parts ...string) DiskPath {
	return DiskPath(filepath.Join(parts...))
}

func (mock *MockDisk) FileExists(path DiskPath) (bool, bool, error) {
	if slices.Contains(mock.Files, string(path)) {
		return true, true, nil
	}
	return false, false, nil
}

func (mock *MockDisk) DirExists(path DiskPath) (bool, bool, error) {
	if slices.Contains(mock.Dirs, string(path)) {
		return true, true, nil
	}
	return false, false, nil
}

func (mock *MockDisk) List(path DiskPath) ([]DiskPath, []DiskPath, error) {
	var files []DiskPath
	var dirs []DiskPath

	for _, file := range mock.Files {
		if file == string(path) {
			continue
		}
		if strings.HasPrefix(file, string(path)) {
			relative, _ := strings.CutPrefix(file, string(path)+"/")
			files = append(files, DiskPath(relative))
		}
	}

	for _, dir := range mock.Dirs {
		if dir == string(path) {
			continue
		}
		if strings.HasPrefix(dir, string(path)) {
			relative, _ := strings.CutPrefix(dir, string(path)+"/")
			dirs = append(dirs, DiskPath(relative))
		}
	}

	return files, dirs, nil
}

func (mock *MockDisk) ArchiveDir(src, dst DiskPath) error {
	return nil
}

func (mock *MockDisk) WriteFile(path DiskPath, data io.Reader) error {
	return nil
}

func (mock *MockDisk) CreateDir(path DiskPath) error {
	mock.Dirs = append(mock.Dirs, string(path))
	return nil
}

func (mock *MockDisk) ListRec(path DiskPath) ([]DiskPath, error) {
	return nil, nil
}

func (mock *MockDisk) CreateTar(path DiskPath) error {
	mock.Files = append(mock.Files, string(path))
	return nil
}

func (mock *MockDisk) Move(toPath DiskPath, fromPath DiskPath, files []DiskPath, overwrite bool, log *internal.Log) bool {
	return false
}

func (mock *MockDisk) Transfer(disk Disk, toPath string, fromPath DiskPath, files []DiskPath, log *internal.Log) bool {
	return false
}

func (mock *MockDisk) CreateDeb(path string, files map[string]DiskPath) error {
	return nil
}

func (mock *MockDisk) ReadFile(path DiskPath) ([]byte, bool, error) {
	return nil, false, nil
}
