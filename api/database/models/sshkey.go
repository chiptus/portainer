package models

import "net/http"

type SSHPassphrase struct {
	SSHPassphrasePayload string `json:"passphrase"`
}

type SSHKeyPair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

func (p *SSHPassphrase) Validate(request *http.Request) error {
	return nil
}

func (p SSHPassphrase) String() string {
	return p.SSHPassphrasePayload
}
