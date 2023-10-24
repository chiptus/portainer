package liblicense

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// AESEncrypt encrypts a plaintext value using the AES algorithm.
func AESEncrypt(plaintext, key []byte) ([]byte, error) {
	// NewCipher creates and returns a new cipher.Block. The key argument should be the AES key, either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// NewGCM returns the given 128-bit, block cipher wrapped in Galois Counter Mode with the standard nonce length
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal encrypts and authenticates plaintext, authenticates the
	// additional data and appends the result to dst, returning the updated
	// slice. The nonce must be NonceSize() bytes long and unique for all
	// time, for a given key.
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// AESDecrypt decrypts an encodedtext value using the AES algorithm.
func AESDecrypt(encodedtext, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encodedtext) < nonceSize {
		return nil, errors.New("encodedtext too short")
	}

	nonce, encodedtext := encodedtext[:nonceSize], encodedtext[nonceSize:]
	return gcm.Open(nil, nonce, encodedtext, nil)
}
