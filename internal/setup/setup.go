package setup

import (
	"bytes"
	"os"
	"strconv"
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

func SetUp(log *internal.Log) {
	prev := log.Stage("setup")
	defer prev()

	err := os.MkdirAll("/etc/catalogue", 0755)
	if err != nil {
		log.Err(err, "failed to create /etc/catalogue")
		return
	}

	cfg := config(log)
	if err != nil {
		return
	}
	aptKeys(log)
	aptRepository(cfg, log)
}

func config(log *internal.Log) *internal.Config {
	existing, err := os.ReadFile(ext.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info(10, "%s does not exists", ext.ConfigPath)
		} else {
			log.Err(err, "failed to read file %s", ext.ConfigPath)
			return nil
		}
	} else {
		log.Info(10, "%s already exists", ext.ConfigPath)
		cfg, err := internal.ParseConfig(bytes.NewReader(existing))
		if err != nil {
			log.Err(err, "failed to parse cong file")
			return nil
		}
		return &cfg
	}

	cfg := internal.DefaultConfig()

	data := bytes.NewBuffer([]byte{})

	err = internal.SerializeConfig(data, cfg)
	if err != nil {
		log.Err(err, "failed to serialize config")
		return nil
	}

	err = os.WriteFile(ext.ConfigPath, data.Bytes(), 0755)
	if err != nil {
		log.Err(err, "failed to create %s", ext.ConfigPath)
	} else {
		log.Info(10, "created %s", ext.ConfigPath)
	}
	return &cfg
}

func aptKeys(log *internal.Log) {

	newPublic := false

	_, err := os.Stat(ext.APTPublicGPGKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info(10, "%s does not exists", ext.APTPublicGPGKeyPath)
			newPublic = true
		} else {
			log.Err(err, "failed to check %s", ext.APTPublicGPGKeyPath)
			return
		}
	} else {
		log.Info(10, "%s already exists", ext.APTPublicGPGKeyPath)
	}

	newPrivate := false
	_, err = os.Stat(ext.APTPrivateGPGKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info(10, "%s does not exists", ext.APTPrivateGPGKeyPath)
			newPrivate = true
		} else {
			log.Err(err, "failed to check %s", ext.APTPrivateGPGKeyPath)
			return
		}
	} else {
		log.Info(10, "%s already exists", ext.APTPrivateGPGKeyPath)
	}

	generate := newPrivate != newPublic || newPrivate
	if newPrivate != newPublic {
		log.Info(10, "did not find both public and private apt repository keys, will regenerate")
	} else if newPrivate {
		log.Info(10, "missing apt repository keys, will generate")
	}

	if !generate {
		return
	}

	err = os.MkdirAll(ext.APTKeyRingPath, 0755)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Err(err, "failed to create %s", ext.APTKeyRingPath)
			return
		}
	}

	keys, err := internal.CreateOpenPGPKey()
	if err != nil {
		log.Err(err, "failed to generate opengpg keys")
		return
	}

	err = os.WriteFile(ext.APTPublicGPGKeyPath, keys.Public, 0755)
	if err != nil {
		log.Err(err, "failed to write public apt repository key to %s", ext.APTPublicGPGKeyPath)
		return
	}
	log.Info(10, "wrote public apt repository key to %s", ext.APTPublicGPGKeyPath)

	err = os.WriteFile(ext.APTPrivateGPGKeyPath, keys.Private, 0755)
	if err != nil {
		log.Err(err, "failed to write public key to %s", ext.APTPrivateGPGKeyPath)
		return
	}
	log.Info(10, "wrote private apt repository key to %s", ext.APTPrivateGPGKeyPath)
	return
}

func aptRepository(config *internal.Config, log *internal.Log) {
	_, err := os.Stat(ext.APTSourceListPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Err(err, "failed to check %s", ext.APTSourceListPath)
			return
		} else {
			log.Info(10, "%s already exists, will regenerate", ext.APTSourceListPath)
		}
	}

	file, err := os.Create(ext.APTSourceListPath)
	if err != nil {
		log.Err(err, "failed to create %s", ext.APTSourceListPath)
		return
	}
	defer file.Close()

	contents := strings.Builder{}
	contents.WriteString("# This file is auto-generated, any changes may be override later on with `catalogue setup`\n")

	contents.WriteString("deb [signed-by=")
	contents.WriteString(ext.APTPublicGPGKeyPath)
	contents.WriteString("] http://localhost:")
	contents.WriteString(strconv.Itoa(config.Port))
	contents.WriteString(" stable packages\n")

	_, err = file.Write([]byte(contents.String()))

	if err != nil {
		log.Err(err, "failed to write apt repository source %s", ext.APTSourceListPath)
		return
	}

	log.Info(10, "created %s", ext.APTSourceListPath)

}
