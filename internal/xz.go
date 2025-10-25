package internal

import (
	"bytes"

	xzlib "github.com/ulikunitz/xz"
)

func XZ(in []byte) ([]byte, error) {

	var compressed bytes.Buffer

	writer, err := xzlib.NewWriter(&compressed)
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	_, err = writer.Write(in)
	if err != nil {
		return nil, err
	}

	return compressed.Bytes(), nil
}
