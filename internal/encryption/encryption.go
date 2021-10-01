package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type Encryption struct {
	// Key is a hash of bytes that instantiates the encryption package
	Key []byte
}

func (e *Encryption) Encrypt(text string) (string, error) {
	plaintext := []byte(text)

	// create block for encryption
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	// create the cypher
	cipherText := make([]byte, aes.BlockSize+len(plaintext))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plaintext)

	// return as base 64
	return base64.URLEncoding.EncodeToString(cipherText), nil
}

func (e *Encryption) Decrypt(crypoText string) (string, error) {
	cipherText, err := base64.RawStdEncoding.DecodeString(crypoText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	// unable to decrypt if cipher len is less than block size
	if len(cipherText) < aes.BlockSize {
		return "", err
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return fmt.Sprintf("%b", cipherText), nil
}
