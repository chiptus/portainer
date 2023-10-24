package liblicense

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
)

// LicenseExpiresAt provides expiration time of the given license
func LicenseExpiresAt(license PortainerLicense) int64 {
	return ExpiresAt(license.Created, license.ExpiresAfter).Unix()
}

// EndOfDay returns end of day time of a given time
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
}

// ExpiresAt calculate expiration time of license
func ExpiresAt(created int64, expiration int) time.Time {
	return EndOfDay(time.Unix(created, 0)).Add(time.Hour * time.Duration(expiration) * 24)
}

// Expired check if a license is expired
func Expired(created int64, expiration int) bool {
	return time.Now().After(ExpiresAt(created, expiration))
}

// ValidateLicense takes a license object
// validates it against the license server
func ValidateLicense(license *PortainerLicense) (bool, error) {
	if license.LicenseKey == "" {
		return false, errors.New("license key is empty")
	}

	if Expired(license.Created, license.ExpiresAfter) {
		return false, errors.New("license validity date has passed")
	}

	// deliberately skip errors to allow usage of a revoked key inside
	// an offline environment.
	valid, err := isValidLicense(license.LicenseKey)
	if err != nil {
		log.Printf("[DEBUG] [liblicense,validate] [msg: Failed to validate license with license server] [err: %s]", err)
	}

	return valid, nil
}

func isValidLicense(licenseKey string) (bool, error) {
	type licenseValidationResponse map[string]bool

	type licenseValidationPayload struct {
		LicenseKeys []string
	}

	jsonPayload := licenseValidationPayload{
		LicenseKeys: []string{licenseKey},
	}

	payload, err := json.Marshal(jsonPayload)
	if err != nil {
		return true, err
	}

	data, err := Post(LicenseCheckURL, payload, 0)
	if err != nil {
		return true, trimError(err)
	}

	var validationResponse licenseValidationResponse
	err = json.Unmarshal(data, &validationResponse)
	if err != nil {
		return true, err
	}

	valid, _ := validationResponse[licenseKey]

	return valid, nil
}

// ignore the possible http request errors in offline env
func trimError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no such host"):
		return nil
	case strings.Contains(msg, "408"): // request timeout
		return nil
	case strings.Contains(msg, "502"): // bad gateway
		return nil
	case strings.Contains(msg, "503"): // service unavailable
		return nil
	case strings.Contains(msg, "504"): // gateway timeout
		return nil
	default:
		return err
	}
}
