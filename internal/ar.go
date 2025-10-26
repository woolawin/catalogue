package internal

import (
	"io"
	"os"
	"time"

	arlib "github.com/blakesmith/ar"
)

func CreateAR(input map[string]string, dst io.Writer) error {

	writer := arlib.NewWriter(dst)
	err := writer.WriteGlobalHeader()
	if err != nil {
		return ErrOf(err, "failed to write ar header")
	}

	for name, path := range input {
		err := addFileToAr(writer, name, string(path), 0644)
		if err != nil {
			return ErrOf(err, "can not add file '%s' to .deb", name)
		}
	}

	return nil
}

func addFileToAr(writer *arlib.Writer, name string, filePath string, mode int64) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ErrOf(err, "can not read file '%s'", filePath)
	}

	header := &arlib.Header{
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
