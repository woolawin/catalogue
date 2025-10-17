package build

import "fmt"

func data(src BuildSrc) error {

	tarPath := filePath(src, "data.tar.gz")
	dirPath := filePath(src, "data")

	exists, asFile, err := fileExists(tarPath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if !asFile {
		return fmt.Errorf("data.tar.gz is not a file")
	}

	exists, asDir, err := dirExists(dirPath)
	if err != nil {
		return err
	}
	if !asDir {
		return fmt.Errorf("data is not a directory")
	}

	if !exists {
		emptyTar(tarPath)
	}

	return archiveDirectory(dirPath, tarPath)

}
