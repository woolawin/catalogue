package internal

import (
	"bytes"
	"crypto"
	"time"

	pgplib "github.com/ProtonMail/go-crypto/openpgp"
	armorlib "github.com/ProtonMail/go-crypto/openpgp/armor"
	packetlib "github.com/ProtonMail/go-crypto/openpgp/packet"
)

func PGPSign(key *pgplib.Entity, data []byte) (string, error) {
	var buf bytes.Buffer
	armored, err := armorlib.Encode(&buf, "PGP SIGNATURE", nil)
	if err != nil {
		return "", err
	}
	defer armored.Close()

	writer, err := pgplib.Sign(armored, key, nil, nil)
	if err != nil {
		return "", err
	}
	_, err = writer.Write(data)
	if err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}

	err = armored.Close()
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

type PGPKey struct {
	Public  []byte
	Private []byte
}

func CreateOpenPGPKey() (*PGPKey, error) {
	name := "Catalogue APT Server"
	email := "catalogue@localhost"
	comment := ""

	config := &packetlib.Config{
		DefaultHash: crypto.SHA256,
		RSABits:     4096,
		Time:        func() time.Time { return time.Now() },
	}

	entity, err := pgplib.NewEntity(name, comment, email, config)
	if err != nil {
		return nil, err
	}

	var public bytes.Buffer
	if err := entity.Serialize(&public); err != nil {
		return nil, err
	}

	var private bytes.Buffer
	if err := entity.SerializePrivate(&private, nil); err != nil {
		return nil, err
	}

	return &PGPKey{Public: public.Bytes(), Private: private.Bytes()}, nil
}

func ReadPrivateKey(private []byte) (*pgplib.Entity, error) {
	entities, err := pgplib.ReadKeyRing(bytes.NewReader(private))
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return nil, err
	}
	return entities[0], nil
}
