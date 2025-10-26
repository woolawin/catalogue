package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

func CreateAR(files []string, dst io.Writer, log *Log) bool {
	prev := log.Stage("deb")
	defer prev()

	err := os.MkdirAll("/tmp/catalogue-ar", 0755)
	if err != nil {
		log.Err(err, "failed to make tmp dir /tmp/catalogue-ar")
		return false
	}

	path := fmt.Sprintf("/tmp/catalogue-ar/%d.ar", time.Now().UnixMilli())

	args := make([]string, len(files)+2)
	args[0] = "rcs"
	args[1] = path
	for idx, file := range files {
		args[idx+2] = file
	}

	ar := exec.Command("ar", args...)
	stdout, err := ar.CombinedOutput()
	if err != nil {
		log.Err(Err(string(stdout)), "failed to archive file %s", "path")
		return false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Err(err, "failed to open file %s", path)
		return false
	}
	// defer file.Close()

	// _, err = io.Copy(dst, file)
	_, err = dst.Write(data)
	if err != nil {
		log.Err(err, "failed to copy ar file %s", path)
		return false
	}

	return true

}
