package liblicense

import (
	"errors"
	"log"
	"strconv"
	"strings"
)

var errInvalidLicense = errors.New("license is invalid")

// ParseLicenseKey takes a license key and parses the key.
func ParseLicenseKey(licenseKey string) (*PortainerLicense, error) {
	if len(licenseKey) < 2 {
		return nil, errInvalidLicense
	}

	var key []byte
	switch {
	case strings.HasPrefix(licenseKey, "2-"):
		key = []byte(MasterKeyV2)
	case strings.HasPrefix(licenseKey, "3-"):
		key = []byte(MasterKeyV3)
	default:
		log.Printf("[DEBUG] [license: %s] [msg: license has invalid prefix]", licenseKey)
		return nil, errInvalidLicense
	}

	encryptedLicense, err := PortainerEncoding.DecodeString(licenseKey[2:])
	if err != nil {
		return nil, errInvalidLicense
	}
	plainTextLicense, err := AESDecrypt(encryptedLicense, key)
	if err != nil {
		return nil, errInvalidLicense
	}

	license, err := licenseFromString(string(plainTextLicense))
	if err != nil {
		return nil, errInvalidLicense
	}

	license.LicenseKey = licenseKey

	return license, nil
}

func licenseFromString(plainTextLicense string) (*PortainerLicense, error) {
	licenseInfo := strings.Split(plainTextLicense, "|")

	if len(licenseInfo) < 7 {
		log.Printf("[DEBUG] [license: %s] [msg: Not enough parts]", plainTextLicense)
		return nil, errInvalidLicense
	}

	productEdition, err := strconv.Atoi(licenseInfo[0])
	if err != nil {
		return nil, err
	}

	licenseType, err := strconv.Atoi(licenseInfo[1])
	if err != nil {
		return nil, err
	}

	generationDate, err := strconv.ParseInt(licenseInfo[2], 10, 64)
	if err != nil {
		return nil, err
	}

	expiresAfter, err := strconv.Atoi(licenseInfo[3])
	if err != nil {
		return nil, err
	}

	company := licenseInfo[4]

	nodes, err := strconv.Atoi(licenseInfo[5])
	if err != nil {
		return nil, err
	}

	version, err := strconv.Atoi(licenseInfo[6])
	if err != nil {
		return nil, err
	}

	return &PortainerLicense{
		Company: company,
		Created: generationDate,
		// license.Email =
		ExpiresAfter:   expiresAfter,
		ExpiresAt:      ExpiresAt(generationDate, expiresAfter).Unix(),
		Nodes:          nodes,
		ProductEdition: ProductEdition(productEdition),
		Type:           PortainerLicenseType(licenseType),
		Version:        version,
	}, nil
}
