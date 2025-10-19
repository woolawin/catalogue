package ext

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/woolawin/catalogue/internal"
)

type Disk interface {
	Path(parts ...string) string
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data io.Reader) error
	FileExists(path string) (bool, bool, error)
	DirExists(path string) (bool, bool, error)
	CreateDir(path string) error
	List(path string) ([]string, []string, error)
	ListRec(path string) ([]string, error)
	CreateTar(path string) error
	ArchiveDir(src, dst string) error
	ArchiveFiles(path string, files []string) error
	Move(toPath string, fromPath string, files []string, overwrite bool) error
}

func NewDisk(base string) Disk {
	return &diskImpl{base: base}
}

type diskImpl struct {
	base string
}

func (disk *diskImpl) Path(parts ...string) string {
	return filepath.Join(slices.Insert(parts, 0, string(disk.base))...)
}

func (disk *diskImpl) ReadFile(path string) ([]byte, error) {
	if disk.unsafe(path) {
		return nil, internal.ErrFileBlocked(path, "read")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, internal.ErrOf(err, "can not read file '%s'", path)
	}

	return data, nil
}

func (disk *diskImpl) WriteFile(path string, data io.Reader) error {
	if disk.unsafe(path) {
		return internal.ErrFileBlocked(path, "read")
	}
	file, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "can not create file %s", path)
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		return internal.ErrOf(err, "failed to write to file %s", path)
	}
	return nil
}

func (disk *diskImpl) FileExists(path string) (bool, bool, error) {
	if disk.unsafe(path) {
		return false, false, internal.ErrFileBlocked(path, "read")
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, true, nil
		}
		return false, false, err
	}
	return true, !info.IsDir(), nil
}

func (disk *diskImpl) DirExists(path string) (bool, bool, error) {
	if disk.unsafe(path) {
		return false, false, internal.ErrFileBlocked(path, "read")
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, true, nil
		}
		return false, false, err
	}
	return true, info.IsDir(), nil
}

func (disk *diskImpl) CreateDir(path string) error {
	if disk.unsafe(path) {
		return internal.ErrFileBlocked(path, "created")
	}
	err := os.Mkdir(path, 0755)
	if err != nil {
		return internal.ErrOf(err, "can not create directory %s", path)
	}
	return nil
}

func (disk *diskImpl) List(path string) ([]string, []string, error) {
	if disk.unsafe(path) {
		return nil, nil, internal.ErrFileBlocked(path, "read")
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, internal.ErrOf(err, "can not list directory %s", path)
	}

	var files []string
	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		} else {
			files = append(files, entry.Name())
		}
	}

	return files, dirs, nil
}

func (disk *diskImpl) ListRec(path string) ([]string, error) {
	if disk.unsafe(path) {
		return nil, internal.ErrFileBlocked(path, "read")
	}
	var files []string
	err := filepath.WalkDir(path, func(entryPath string, entry os.DirEntry, err error) error {
		if err != nil {
			return internal.ErrOf(err, "can not list directory %s", path)
		}
		relative, _ := strings.CutPrefix(entryPath, path+"/")
		if !entry.IsDir() {
			files = append(files, relative)
		}
		return nil
	})

	if err != nil {
		return nil, internal.ErrOf(err, "can not recusrivly list directory %s", path)
	}
	return files, nil
}

func (disk *diskImpl) CreateTar(path string) error {
	if disk.unsafe(path) {
		return internal.ErrFileBlocked(path, "created")
	}
	file, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "can not create file data.tar.gz")
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	return nil
}

func (disk *diskImpl) Move(toPath string, fromPath string, files []string, overwrite bool) error {
	if disk.unsafe(toPath) {
		return internal.ErrFileBlocked(toPath, "written")
	}
	for _, file := range files {
		newPath := filepath.Join(toPath, file)
		if disk.unsafe(newPath) {
			return internal.ErrFileBlocked(toPath, "written")
		}
		oldPath := filepath.Join(fromPath, file)
		if disk.unsafe(oldPath) {
			return internal.ErrFileBlocked(oldPath, "read")
		}
		_, err := os.Stat(newPath)
		if err == nil {
			continue
		} else if !os.IsNotExist(err) {
			continue
		}
		os.MkdirAll(filepath.Dir(newPath), 0755)
		err = os.Rename(oldPath, newPath)
		if err != nil {
			return internal.ErrOf(err, "can not move file %s to %s", fromPath, toPath)
		}
	}

	return nil
}

func (disk *diskImpl) ArchiveDir(src string, dst string) error {
	if disk.unsafe(src) {
		return internal.ErrFileBlocked(src, "read")
	}
	if disk.unsafe(dst) {
		return internal.ErrFileBlocked("dst", "written")
	}

	file, err := os.Create(dst)
	if err != nil {
		return internal.ErrOf(err, "can not create file %s", dst)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return internal.ErrOf(err, "can not list archive file %s", file)
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return internal.ErrOf(err, "can not read file header %s", file)
		}

		relPath, err := filepath.Rel(filepath.Dir(src), file)
		if err != nil {
			return internal.ErrOf(err, "can not determine relative file for %s", file)
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return internal.ErrOf(err, "can not write header for file %s", file)
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return internal.ErrOf(err, "can not open archive file %s", file)
		}
		defer f.Close()

		_, err = io.Copy(tarWriter, f)
		if err != nil {
			return internal.ErrOf(err, "can not write archive file %s", file)
		}
		return nil
	})
	if err != nil {
		return internal.ErrOf(err, "can not archive directory %s from %s", dst, src)
	}
	return nil
}

func (disk *diskImpl) ArchiveFiles(path string, files []string) error {
	if disk.unsafe(path) {
		return internal.ErrFileBlocked(path, "written")
	}

	archive, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "can not create archive '%s'", path)
	}
	defer archive.Close()

	for _, file := range files {
		if disk.unsafe(file) {
			return internal.ErrFileBlocked(file, "read")
		}
		if err := addFile(archive, file); err != nil {
			return internal.ErrOf(err, "can not add file '%s' to archive '%s'", file, path)
		}
	}
	return nil
}

func addFile(writer io.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return internal.ErrOf(err, "can not open file")
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return internal.ErrOf(err, "can not stat file")
	}

	name := filepath.Base(path)
	modTime := info.ModTime().Unix()
	size := info.Size()
	mode := info.Mode().Perm()

	// Create header (fixed width fields)
	// Format: (each field is padded with spaces)
	// 0-15  : File identifier
	// 16-27 : File modification timestamp
	// 28-33 : Owner ID
	// 34-39 : Group ID
	// 40-47 : File mode
	// 48-57 : File size in bytes
	// 58-59 : Trailer "`\n"
	header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8o%-10d`\n",
		name,
		modTime,
		0,    // UID
		0,    // GID
		mode, // File mode
		size, // File size
	)

	if _, err := writer.Write([]byte(header)); err != nil {
		return err
	}

	if _, err := io.Copy(writer, file); err != nil {
		return internal.ErrOf(err, "can not write bytes to file")
	}

	if size%2 != 0 {
		if _, err := writer.Write([]byte("\n")); err != nil {
			return internal.ErrOf(err, "can not write newline to file")
		}
	}

	return nil
}
func (disk *diskImpl) unsafe(path string) bool {
	baseAbs, err := filepath.Abs(disk.base)
	if err != nil {
		return true
	}

	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return true
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
		return true
	}
	return false
}
