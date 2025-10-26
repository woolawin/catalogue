package deb

import (
	"os"
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

func BuildDebFile(in string, outPath string, log *internal.Log) bool {
	disk := ext.NewDisk(in)

	err := disk.WriteFile(disk.Path("debian-binary"), strings.NewReader("2.0\n"))
	if err != nil {
		log.Err(err, "failed to write debian-binary file")
		return false
	}

	err = disk.ArchiveDir(disk.Path("data"), disk.Path("data.tar.gz"))
	if err != nil {
		log.Err(err, "failed to create data.tar.gz archive")
		return false
	}

	err = disk.ArchiveDir(disk.Path("control"), disk.Path("control.tar.gz"))
	if err != nil {
		log.Err(err, "failed to create control.tar.gz archive")
		return false
	}

	files := map[string]string{
		"debian-binary":  string(disk.Path("debian-binary")),
		"control.tar.gz": string(disk.Path("control.tar.gz")),
		"data.tar.gz":    string(disk.Path("data.tar.gz")),
	}

	file, err := os.Create(outPath)
	if err != nil {
		log.Err(err, "failed to open %s", outPath)
		return false
	}
	defer file.Close()

	err = internal.CreateAR(files, file)
	if err != nil {
		log.Err(err, "failed to create archive %s", outPath)
		return false
	}
	return true
}
