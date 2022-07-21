package cryprot_test

import (
	"reflect"
	"testing"

	"github.com/crn4/yameee/cryprot"
)

func TestKeyPairs(t *testing.T) {
	publicKeyPeer1, privateKeyPeer1 := cryprot.GeneratePubPrivKeys()
	publicKeyPeer2, privateKeyPeer2 := cryprot.GeneratePubPrivKeys()

	secretKeyPeer1 := cryprot.CalcSecretKey(publicKeyPeer2, privateKeyPeer1)
	secretKeyPeer2 := cryprot.CalcSecretKey(publicKeyPeer1, privateKeyPeer2)

	if !reflect.DeepEqual(secretKeyPeer1, secretKeyPeer2) {
		t.Fatal("secretKey1 is not equal to secretKey2")
	}

	message := []byte("Do not go gentle into that good night,	Old age should burn and rave at close of day;	Rage, rage against the dying of the light.")
	cipherMessage := cryprot.EncryptMessage(message, secretKeyPeer1)
	decryptedMessage := cryprot.DecryptMessage(cipherMessage, secretKeyPeer2)

	if !reflect.DeepEqual(message, decryptedMessage) {
		t.Fatal("decrypted message is not equal to original message")
	}
}
