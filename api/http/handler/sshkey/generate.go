package sshkey

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/database/models"
	"golang.org/x/crypto/ssh"
)

const BitSize = 4096

// @id Generate
// @summary Generate ssh keypair
// @description Generate an ssh public / private keypair
// @description **Access policy**: authenticated
// @tags cloud_credentials
// @security ApiKeyAuth
// @security jwt
// @accept json,multipart/form-data
// @produce json
// @success 200 {object} models.SSHKeyPair
// @failure 500 "Server error"
// @router /sshkeygen [post]
func (h *Handler) generate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var passphrase models.SSHPassphrase
	err := request.DecodeAndValidateJSONPayload(r, &passphrase)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	privateKey, err := generatePrivateKey(BitSize)
	if err != nil {
		return httperror.InternalServerError("unable to generate private key", err)
	}
	publicKeyBytes, err := generatePublicKey(&privateKey.PublicKey)
	if err != nil {
		return httperror.InternalServerError("unable to generate public key", err)
	}

	privateKeyBytes, err := encodePrivateKeyToPEM(privateKey, passphrase.String())
	if err != nil {
		return httperror.InternalServerError("unable to encrypt privatekey", err)
	}

	keyPair := models.SSHKeyPair{
		Private: string(privateKeyBytes),
		Public:  string(publicKeyBytes),
	}
	return response.JSON(w, keyPair)
}

// generatePrivateKey creates a RSA Private Key of specified byte size.
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// generatePublicKey generates a matching public key from an RSA privatekey.
func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)
	return pubKeyBytes, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format.
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey, passphrase string) ([]byte, error) {
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if passphrase != "" {
		var err error
		block, err = x509.EncryptPEMBlock(
			rand.Reader,
			block.Type,
			block.Bytes,
			[]byte(passphrase),
			x509.PEMCipherAES256,
		)
		if err != nil {
			return nil, err
		}
	}

	return pem.EncodeToMemory(block), nil
}
