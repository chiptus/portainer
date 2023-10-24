package liblicense

import "encoding/base64"

var (
	// LicenseServerBaseURL represents the base URL of the API used to validate
	// an extension license.
	LicenseServerBaseURL = "https://api.portainer.io"
	// ExtensionLicenseCheckURL represents the URL of the API used to validate
	// an extension license.
	ExtensionLicenseCheckURL = LicenseServerBaseURL + "/license/validate"
	// LicenseCheckURL represents the URL of the API used to validate an
	// extension license.
	LicenseCheckURL = LicenseServerBaseURL + "/licenses/validate"
	// PortainerEncoding is the encoding used to encode and decode licenses
	// keys.
	PortainerEncoding = base64.StdEncoding
)

const (
	// MasterKeyV2 represents the encryption key used in Version 1/2 licenses.
	// A prefix of 1 or 2 before a license indicates that it should be
	// decrypted with this key.
	MasterKeyV2 = `portaineriomasterkey!!12.12.2018`

	// MasterKeyV3 represents the encryption key used in Version 3 licenses.
	// A prefix of 3 before a license indicates that it should be decrypted
	// with this key.
	MasterKeyV3 = `?-Dez$&+4(!y|FAFx;O<"'R=n{=J*es"`
)

type (
	// PortainerLicense represents all details of a portainer license stored in
	// our license server. NOTE: Not all of these details are encoded within
	// the license key itself. See generation.go for which fields are currently
	// encoded.
	PortainerLicense struct {
		ID           string `json:"id,omitempty"`
		Company      string `json:"company,omitempty"`
		Created      int64  `json:"created,omitempty"`
		Email        string `json:"email,omitempty"`
		ExpiresAfter int    `json:"expiresAfter,omitempty"`
		LicenseKey   string `json:"licenseKey,omitempty"`
		Nodes        int    `json:"nodes,omitempty"`

		// ProductEdition was created originally with plans on having a
		// seperate portainer product for Enterprise users and Business users
		// with differing features. This didn't wind up coming about, but may
		// still serve useful in the future if we need to issue keys for a
		// different product entirely.
		// Originally, the ProductEdition was used as the prefix for generating
		// license keys, but in practice most people thought it was the
		// "version" due to us having the original extension licenses which can
		// be thought of as the true version 1 licenses.
		ProductEdition ProductEdition `json:"productEdition,omitempty"`
		Revoked        bool           `json:"revoked,omitempty"`
		RevokedAt      int64          `json:"revokedAt,omitempty"`

		// Type is used to distinguish different kinds of licenses, trial
		// licenses, enterprise subscriptions
		Type PortainerLicenseType `json:"type,omitempty"`

		// Version indicates which key should be used to encode/decode the
		// license string.
		Version      int    `json:"version,omitempty"`
		Reference    string `json:"reference,omitempty"`
		ExpiresAt    int64  `json:"expiresAt,omitempty"`
		FirstCheckin int64  `json:"firstCheckin,omitempty"`
		LastCheckin  int64  `json:"lastCheckin,omitempty"`
		UniqueId     string `json:"uniqueId,omitempty" gorm:"primaryKey"`
		RedisRef     string `json:"redisRef,omitempty"`
	}

	// PortainerLicenseType represents an enum for types of Portainer License.
	PortainerLicenseType int

	// ProductEdition was created originally with plans on having a seperate
	// portainer product for Enterprise users and Business users with differing
	// features. This didn't wind up coming about, but may still serve useful
	// in the future if we need to issue keys for a different product entirely.
	ProductEdition int
)

// PortainerLicenseType represents an enum for types of Portainer License
const (
	_ PortainerLicenseType = iota
	PortainerLicenseTrial
	PortainerLicenseSubscription
	PortainerLicenseFree
	PortainerLicensePersonal
	PortainerLicenseStarter
)

const (
	_ ProductEdition = iota
	// PortainerCE represents the community edition of Portainer.
	// NOTE: This edition is not used, but exists for completion sake.
	PortainerCE
	// PortainerBE represents the business edition of Portainer.
	// NOTE: This is the only edition in actual use. All existing keys will
	// have a a ProductEdition of 2.
	PortainerBE
	// PortainerEE represents the enterprise edition of Portainer.
	// NOTE: This edition does not exist, but perhaps someday we'll create
	// something like it.
	PortainerEE
)
