package internal

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func TestSignPGP(t *testing.T) {
	key, err := CreateOpenPGPKey()
	if err != nil {
		t.Fatal(err)
	}

	priv, err := ReadPrivateKey(key.Private)
	if err != nil {
		t.Fatal(err)
	}

	message := make([]byte, 64)
	_, err = rand.Read(message)
	if err != nil {
		t.Fatal(err)
	}

	signature, err := PGPSign(priv, message)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(signature))
}
