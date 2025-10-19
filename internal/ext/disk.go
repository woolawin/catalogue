package ext

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/blakesmith/ar"
	"github.com/woolawin/catalogue/internal"
)

type DiskPath string

type Disk interface {
	Path(parts ...string) DiskPath
	ReadFile(path DiskPath) ([]byte, error)
	WriteFile(path DiskPath, data io.Reader) error
	FileExists(path DiskPath) (bool, bool, error)
	DirExists(path DiskPath) (bool, bool, error)
	CreateDir(path DiskPath) error
	List(path DiskPath) ([]DiskPath, []DiskPath, error)
	ListRec(path DiskPath) ([]DiskPath, error)
	CreateTar(path DiskPath) error
	ArchiveDir(src, dst DiskPath) error
	CreateDeb(path string, files map[string]DiskPath) error
	Move(toPath DiskPath, fromPath DiskPath, files []DiskPath, overwrite bool) error
}

func NewDisk(base string) Disk {
	return &diskImpl{base: base}
}

type diskImpl struct {
	base string
}

func (disk *diskImpl) Path(parts ...string) DiskPath {
	return DiskPath(filepath.Join(slices.Insert(parts, 0, string(disk.base))...))
}

func (disk *diskImpl) ReadFile(path DiskPath) ([]byte, error) {
	if disk.unsafe(path) {
		return nil, internal.ErrFileBlocked(string(path), "read")
	}
	data, err := os.ReadFile(string(path))
	if err != nil {
		return nil, internal.ErrOf(err, "can not read file '%s'", path)
	}

	return data, nil
}

func (disk *diskImpl) WriteFile(path DiskPath, data io.Reader) error {
	if disk.unsafe(path) {
		return internal.ErrFileBlocked(string(path), "read")
	}
	file, err := os.Create(string(path))
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

func (disk *diskImpl) FileExists(path DiskPath) (bool, bool, error) {
	if disk.unsafe(path) {
		return false, false, internal.ErrFileBlocked(string(path), "read")
	}
	info, err := os.Stat(string(path))
	if err != nil {
		if os.IsNotExist(err) {
			return false, true, nil
		}
		return false, false, err
	}
	return true, !info.IsDir(), nil
}

func (disk *diskImpl) DirExists(path DiskPath) (bool, bool, error) {
	if disk.unsafe(path) {
		return false, false, internal.ErrFileBlocked(string(path), "read")
	}
	info, err := os.Stat(string(path))
	if err != nil {
		if os.IsNotExist(err) {
			return false, true, nil
		}
		return false, false, err
	}
	return true, info.IsDir(), nil
}

func (disk *diskImpl) CreateDir(path DiskPath) error {
	if disk.unsafe(path) {
		return internal.ErrFileBlocked(string(path), "created")
	}
	err := os.Mkdir(string(path), 0755)
	if err != nil {
		return internal.ErrOf(err, "can not create directory %s", path)
	}
	return nil
}

func (disk *diskImpl) List(path DiskPath) ([]DiskPath, []DiskPath, error) {
	if disk.unsafe(path) {
		return nil, nil, internal.ErrFileBlocked(string(path), "read")
	}
	entries, err := os.ReadDir(string(path))
	if err != nil {
		return nil, nil, internal.ErrOf(err, "can not list directory %s", path)
	}

	var files []DiskPath
	var dirs []DiskPath

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, DiskPath(entry.Name()))
		} else {
			files = append(files, DiskPath(entry.Name()))
		}
	}

	return files, dirs, nil
}

func (disk *diskImpl) ListRec(path DiskPath) ([]DiskPath, error) {
	if disk.unsafe(path) {
		return nil, internal.ErrFileBlocked(string(path), "read")
	}
	var files []DiskPath
	err := filepath.WalkDir(string(path), func(entryPath string, entry os.DirEntry, err error) error {
		if err != nil {
			return internal.ErrOf(err, "can not list directory %s", path)
		}
		relative, _ := strings.CutPrefix(entryPath, string(path)+"/")
		if !entry.IsDir() {
			files = append(files, DiskPath(relative))
		}
		return nil
	})

	if err != nil {
		return nil, internal.ErrOf(err, "can not recusrivly list directory %s", path)
	}
	return files, nil
}

func (disk *diskImpl) CreateTar(path DiskPath) error {
	if disk.unsafe(path) {
		return internal.ErrFileBlocked(string(path), "created")
	}
	file, err := os.Create(string(path))
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

func (disk *diskImpl) Move(toPath DiskPath, fromPath DiskPath, files []DiskPath, overwrite bool) error {
	if disk.unsafe(toPath) {
		return internal.ErrFileBlocked(string(toPath), "written")
	}
	for _, file := range files {
		newPath := filepath.Join(string(toPath), string(file))
		if disk.unsafe(DiskPath(newPath)) {
			return internal.ErrFileBlocked(string(toPath), "written")
		}
		oldPath := filepath.Join(string(fromPath), string(file))
		if disk.unsafe(DiskPath(oldPath)) {
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

func (disk *diskImpl) ArchiveDir(src DiskPath, dst DiskPath) error {
	if disk.unsafe(src) {
		return internal.ErrFileBlocked(string(src), "read")
	}
	if disk.unsafe(dst) {
		return internal.ErrFileBlocked("dst", "written")
	}

	file, err := os.Create(string(dst))
	if err != nil {
		return internal.ErrOf(err, "can not create file %s", dst)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(string(src), func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return internal.ErrOf(err, "can not list archive file %s", file)
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return internal.ErrOf(err, "can not read file header %s", file)
		}

		relPath, err := filepath.Rel(filepath.Dir(string(src)), file)
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

func (impl *diskImpl) CreateDeb(path string, input map[string]DiskPath) error {
	out, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "can not create .deb file", path)
	}
	defer out.Close()

	arWriter := ar.NewWriter(out)
	if err := arWriter.WriteGlobalHeader(); err != nil {
		return internal.ErrOf(err, "failed to write ar header")
	}

	for name, path := range input {
		if impl.unsafe(path) {
			return internal.ErrFileBlocked(string(path), "copied")
		}
		err := addFileToAr(arWriter, name, string(path), 0644)
		if err != nil {
			return internal.ErrOf(err, "can not add file '%s' to .deb", name)
		}
	}

	return nil
}

func addFileToAr(writer *ar.Writer, name string, filePath string, mode int64) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return internal.ErrOf(err, "can not read file '%s'", filePath)
	}

	header := &ar.Header{
		Name:    name,
		ModTime: time.Now().UTC(),
		Uid:     0,
		Gid:     0,
		Mode:    mode,
		Size:    int64(len(data)),
	}

	err = writer.WriteHeader(header)
	if err != nil {
		return internal.ErrOf(err, "can not write header for file '%s'", name)
	}

	_, err = writer.Write(data)
	if err != nil {
		return internal.ErrOf(err, "can not write file '%s'", name)
	}

	return nil
}

func (disk *diskImpl) unsafe(path DiskPath) bool {
	baseAbs, err := filepath.Abs(disk.base)
	if err != nil {
		return true
	}

	pathAbs, err := filepath.Abs(string(path))
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
