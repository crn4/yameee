package cryprot

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"log"

	"golang.org/x/crypto/curve25519"
)

// GeneratePubPrivKey returns public and private key based on Ed25519 algorithm
func GeneratePubPrivKeys() ([]byte, []byte) {
	var publicKey, privateKey = make([]byte, 32), make([]byte, 32)
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalf("%s - %s\n", err, "pub and priv keys were not generated for a peer")
	}

	copy(publicKey[:], pubKey[:])
	copy(privateKey[:], privKey[:])

	publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		log.Fatalf("%s - %s\n", err, "problems with X25519 of publicKey")
	}

	return publicKey, privateKey
}

func CalcSecretKey(pubKey, privKey []byte) (secretKey []byte) {
	secretKey, err := curve25519.X25519(privKey[:], pubKey[:])
	if err != nil {
		log.Fatalf("%s - %s\n", err, "secretKey was not calcullated")
	}
	return secretKey
}

func EncryptMessage(message []byte, secretKey []byte) []byte {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		log.Fatalf("%s - %s\n", err, "aes cipher for a secretKey was not generated (during encryption)")
	}
	cipherMessage := make([]byte, aes.BlockSize+len(message))
	iv := cipherMessage[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatalf("%s - %s\n", err, "the inicialisation vector was not properly filled")
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherMessage[aes.BlockSize:], message)

	return cipherMessage
}

func DecryptMessage(cipherMessage []byte, secretKey []byte) []byte {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		log.Fatalf("%s - %s\n", err, "aes cipfer for a secretKey was not generated (during decription)")
	}
	if len(cipherMessage) < aes.BlockSize {
		log.Fatalf("%s - %s\n", err, "cipherMessage is shorter than aes.BlockSize")
	}
	iv := cipherMessage[:aes.BlockSize]
	cipherMessage = cipherMessage[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherMessage, cipherMessage)

	return cipherMessage
}
