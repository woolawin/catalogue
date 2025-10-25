package setup

import (
	"os"

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

	config(log)
	aptKeys(log)
}

func config(log *internal.Log) {
	_, err := os.Stat(ext.ConfigPath)
	if err == nil {
		log.Info(10, "%s already exists", ext.ConfigPath)
		return
	}
	if err != nil {
		if os.IsNotExist(err) {
			log.Info(10, "%s does not exists", ext.ConfigPath)
		} else {
			log.Err(err, "failed to check %s", ext.ConfigPath)
		}
		return
	}

	defaultConfig := `
default_user = ''
`
	err = os.WriteFile(ext.ConfigPath, []byte(defaultConfig), 0755)
	if err != nil {
		log.Err(err, "failed to create %s", ext.ConfigPath)
	} else {
		log.Info(10, "created %s", ext.ConfigPath)
	}
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
