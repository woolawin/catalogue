package registry

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
)

const releasesCacheBase = "/tmp/catalogue/releases"
const packagesBase = "/var/lib/catalogue/components/packages"

func RemovePackage(name string) (bool, error) {
	err := os.Remove(packagePath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ListPackages() ([]string, error) {
	entries, err := os.ReadDir(packagesBase)
	if err != nil {
		return nil, internal.ErrOf(err, "can not list directory '%s'", packagesBase)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

func CacheRelease(compression string, contents []byte) error {
	if len(contents) == 0 {
		return nil
	}
	path := releaseCachePath(compression, "latest")
	parent := filepath.Dir(path)

	err := os.MkdirAll(parent, 0755)
	if err != nil {
		return internal.ErrOf(err, "can not create releases cache directory '%s'", parent)
	}

	file, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "failed to create release cache file '%s'", path)
	}
	defer file.Close()

	_, err = io.Copy(file, bytes.NewReader(contents))
	if err != nil {
		return internal.ErrOf(err, "can not write to file '%s'", path)
	}

	return nil
}

func ReadReleaseCache(compression string) (string, bool, error) {
	path := releaseCachePath(compression, "latest")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	return string(data), true, nil
}

func GetPackageRecord(packageName string) (config.Record, bool, error) {
	path := packagePath(packageName, "record.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config.Record{}, false, nil
		}
		return config.Record{}, false, internal.ErrOf(err, "can not read file '%s'", path)
	}
	record, err := config.DeserializeRecord(bytes.NewReader(data))
	if err != nil {
		return config.Record{}, false, internal.ErrOf(err, "can not parse package record")
	}
	return record, true, nil
}

func AddPackage(record config.Record) error {
	exists, err := HasPackage(record.Name)
	if err != nil {
		return internal.ErrOf(err, "failed tocheck if package '%s' already exists", record.Name)
	}
	if exists {
		return internal.Err("package '%s' already exists", record.Name)
	}

	err = WriteRecord(record)
	if err != nil {
		return err
	}

	return nil
}

func PackageBuildFile(record config.Record, hash string) (*os.File, error) {
	path := packagePath(record.Name, "caches", hash, "build.deb")
	parent := filepath.Dir(path)
	err := os.MkdirAll(parent, 0755)
	if err != nil {
		return nil, internal.ErrOf(err, "can not create directory '%s'", parent)
	}
	return os.Create(path)
}

func WriteRecord(record config.Record) error {
	path := packagePath(record.Name, "record.toml")

	parent := filepath.Dir(path)
	err := os.MkdirAll(parent, 0755)
	if err != nil {
		return internal.ErrOf(err, "can not create component directory '%s'", parent)
	}

	file, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "can not create component file '%s'", path)
	}
	defer file.Close()

	var buffer bytes.Buffer
	err = config.SerializeRecord(&buffer, record)
	if err != nil {
		return internal.ErrOf(err, "can not serialize record file")
	}

	_, err = io.Copy(file, &buffer)
	if err != nil {
		return internal.ErrOf(err, "can not write to file '%s'", path)
	}

	return nil
}

func HasPackage(name string) (bool, error) {
	path := packagePath(name, "record.toml")
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, internal.ErrOf(err, "can not check if package '%s' exists", name)
	}
	return true, nil
}

func packagePath(parts ...string) string {
	return filepath.Join(append([]string{packagesBase}, parts...)...)
}

func releaseCachePath(parts ...string) string {
	return filepath.Join(append([]string{releasesCacheBase}, parts...)...)
}
