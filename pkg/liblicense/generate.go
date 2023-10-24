package liblicense

import (
	"fmt"
	"strconv"
)

// GenerateLicense generates a new portainer license key.
// There are multiple versions of product keys which are encrypted with
// different master keys. The version field you give determins the version
// which will be produced.
// The plain text format of a license string is the following:
// PRODUCT_EDITION|TYPE|GENERATION_DATE|DAYS|COMPANY|NODES|VERSION
// The key will then be prefixed with a number. A 2 indicates version 1 or 2
// and a 3 indicated version 3.
func GenerateLicense(license *PortainerLicense) (string, error) {
	plainTextLicense := licenseToString(license)

	encryptedLicense, err := encryptLicense(plainTextLicense, license.Version)
	if err != nil {
		return "", err
	}

	return base64Encode(license.Version, encryptedLicense), nil
}

// licenseToString returns license details in the following format:
// PRODUCT_EDITION|TYPE|GENERATION_DATE|DAYS|COMPANY|NODES|VERSION
func licenseToString(license *PortainerLicense) string {
	return fmt.Sprintf(
		"%d|%d|%s|%s|%s|%d|%d",
		license.ProductEdition,
		license.Type,
		strconv.FormatInt(license.Created, 10),
		strconv.Itoa(license.ExpiresAfter),
		license.Company,
		license.Nodes,
		license.Version,
	)
}

// encryptLicense will use the correct key for the given version to encrypt the
// license.
func encryptLicense(license string, version int) ([]byte, error) {
	key := []byte(MasterKeyV3)
	if version <= 2 {
		key = []byte(MasterKeyV2)
	}
	return AESEncrypt([]byte(license), key)
}

func base64Encode(version int, encryptedLicense []byte) string {
	// This is a workaround due to the fact that version 1 licenses had a
	// prefix of 2. No other versions should need workarounds.
	if version == 1 {
		version = 2
	}
	return fmt.Sprintf(
		"%d-%s",
		version,
		PortainerEncoding.EncodeToString(encryptedLicense),
	)
}
