package build

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/woolawin/catalogue/internal/pkge"
)

type BuildSrc string

func Build(src BuildSrc, dst string) error {

	index, err := readPkgeIndex(src)
	if err != nil {
		return err
	}

	err = debianBinary(src)
	if err != nil {
		return err
	}

	err = data(src)
	if err != nil {
		return err
	}

	err = control(src, index)
	if err != nil {
		return err
	}
	return nil
}

func readPkgeIndex(src BuildSrc) (pkge.Index, error) {
	path := filePath(src, "index.catalogue.toml")
	exists, asFile, err := fileExists(path)
	if err != nil {
		return pkge.EmptyIndex(), err
	}

	if !asFile {
		return pkge.EmptyIndex(), fmt.Errorf("index.catalogue.toml is not a file")
	}

	if !exists {
		return pkge.EmptyIndex(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return pkge.EmptyIndex(), fmt.Errorf("could not read index.catalogue.toml: %w", err)
	}

	return pkge.Parse(bytes.NewReader(data))
}

func fileExists(path string) (bool, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, false, err
	}
	return true, !info.IsDir(), nil
}

func dirExists(path string) (bool, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, false, err
	}
	return true, info.IsDir(), nil
}

func archiveDirectory(src string, dst string) error {
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

func emptyTar(path string) error {
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

func filePath(src BuildSrc, parts ...string) string {
	return filepath.Join(slices.Insert(parts, 0, string(src))...)
}

func createDir(path string) error {
	return os.Mkdir(path, 0755)
}
