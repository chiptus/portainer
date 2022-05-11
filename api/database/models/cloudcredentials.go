package models

import (
	"errors"
	"fmt"
	"net/http"
)

type CloudCredentialID int

type CloudCredentialMap map[string]string

type CloudCredential struct {
	ID          CloudCredentialID  `json:"id" example:"1"`
	Provider    string             `json:"provider" example:"aws"`
	Name        string             `json:"name" example:"test-env"`
	Credentials CloudCredentialMap `json:"credentials"`
	Created     int64              `json:"created" example:"1650000000"`
}

func (cr *CloudCredential) Validate(request *http.Request) error {
	if cr.Name == "" {
		return errors.New("missing kubernetes cluster name from the request payload")
	}
	if cr.Provider == "" {
		return errors.New("missing cloud provider from the request payload")
	}

	if request.Method == "POST" {
		if cr.Credentials == nil || len(cr.Credentials) == 0 {
			return errors.New("missing cloud credentials from the request payload")
		}
	}
	return nil
}

func (cr *CloudCredential) ValidateUniqueNameByProvider(cloudCredentials []CloudCredential) error {
	for _, cred := range cloudCredentials {
		// exclude the current credential from the check
		if cr.ID == cred.ID {
			continue
		}
		if cred.Provider == cr.Provider && cr.Name == cred.Name {
			return fmt.Errorf("a credential with this name already exists for %v", cr.Provider)
		}
	}
	return nil
}
