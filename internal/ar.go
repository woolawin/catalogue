package internal

import (
	"io"
	"os"
	"time"

	"github.com/blakesmith/ar"
)

func CreateAR(input map[string]string, writer io.Writer) error {

	arWriter := ar.NewWriter(writer)
	if err := arWriter.WriteGlobalHeader(); err != nil {
		return ErrOf(err, "failed to write ar header")
	}

	for name, path := range input {
		err := addFileToAr(arWriter, name, string(path), 0644)
		if err != nil {
			return ErrOf(err, "can not add file '%s' to .deb", name)
		}
	}

	return nil
}

func addFileToAr(writer *ar.Writer, name string, filePath string, mode int64) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ErrOf(err, "can not read file '%s'", filePath)
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
		return ErrOf(err, "can not write header for file '%s'", name)
	}

	_, err = writer.Write(data)
	if err != nil {
		return ErrOf(err, "can not write file '%s'", name)
	}

	return nil
}
