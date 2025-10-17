package api

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Disk interface {
	Path(parts ...string) string
	FileExists(path string) (bool, bool, error)
	DirExists(path string) (bool, bool, error)
	CreateDir(path string) error
	List(path string) ([]string, []string, error)
	ListRec(path string) ([]string, error)
	CreateTar(path string) error
	Archive(src, dst string) error
}

func NewDisk(base string) Disk {
	return &DiskImpl{base: base}
}

type DiskImpl struct {
	base string
}

func (disk *DiskImpl) Path(parts ...string) string {
	return filepath.Join(slices.Insert(parts, 0, string(disk.base))...)
}

func (disk *DiskImpl) FileExists(path string) (bool, bool, error) {
	err := disk.safe(path)
	if err != nil {
		return false, false, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, false, err
	}
	return true, !info.IsDir(), nil
}

func (disk *DiskImpl) DirExists(path string) (bool, bool, error) {
	err := disk.safe(path)
	if err != nil {
		return false, false, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, false, err
	}
	return true, info.IsDir(), nil
}

func (disk *DiskImpl) CreateDir(path string) error {
	err := disk.safe(path)
	if err != nil {
		return err
	}
	return os.Mkdir(path, 0755)
}

func (disk *DiskImpl) List(path string) ([]string, []string, error) {
	err := disk.safe(path)
	if err != nil {
		return nil, nil, err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read directory contents: %w", err)
	}

	var files []string
	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			files = append(files, entry.Name())
		} else {
			dirs = append(dirs, entry.Name())
		}
	}

	return files, dirs, nil
}

func (disk *DiskImpl) ListRec(path string) ([]string, error) {
	err := disk.safe(path)
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.WalkDir(path, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return files, nil
}

func (disk *DiskImpl) CreateTar(path string) error {
	err := disk.safe(path)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("can not create data.tar.gz: %w", err)
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	return nil
}

func (disk *DiskImpl) Archive(src string, dst string) error {
	err := disk.safe(src)
	if err != nil {
		return err
	}
	err = disk.safe(dst)
	if err != nil {
		return err
	}

	file, err := os.Create(dst)
	if err != nil {
		return nil
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(filepath.Dir(src), file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tarWriter, f)
		return err
	})
}

func (disk *DiskImpl) safe(path string) error {
	baseAbs, err := filepath.Abs(disk.base)
	unsafe := fmt.Errorf("file access not allowed: '%s'", path)
	if err != nil {
		return unsafe
	}

	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return unsafe
	}

	baseAbs = filepath.Clean(baseAbs)
	pathAbs = filepath.Clean(pathAbs)

	if !strings.HasSuffix(baseAbs, string(filepath.Separator)) {
		baseAbs += string(filepath.Separator)
	}

	if !strings.HasSuffix(baseAbs, string(filepath.Separator)) {
		baseAbs += string(filepath.Separator)
	}

	if !strings.HasPrefix(pathAbs, baseAbs) {
		return unsafe
	}
	return nil
}
