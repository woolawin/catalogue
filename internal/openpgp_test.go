package internal

import (
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

	message := "Hello World"
	signature, err := PGPSign(priv, []byte(message))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("============================")
	fmt.Println(string(signature))
}
