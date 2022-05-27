package gke

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Key struct {
	Bytes     []byte
	ProjectID string
}

// extractKey decodes a base64 encoded apiKey and returns the key itself as
// well as an extracted projectID or an error which occured.
func ExtractKey(apiKey string) (Key, error) {
	var k Key

	bytes, err := base64.StdEncoding.DecodeString(apiKey)
	if err != nil {
		return k, fmt.Errorf("GKE api key appears to be invalid: %w", err)
	}
	k.Bytes = bytes

	projectID, err := parseProjectID(bytes)
	if err != nil {
		return k, fmt.Errorf("GKE api key appears to be invalid: %w", err)
	}
	k.ProjectID = projectID
	return k, nil
}

// parseProjectID reads a GKE json keyfile and attempts to extract the
// project_id value.
func parseProjectID(src []byte) (string, error) {
	r := bytes.NewReader(src)
	dec := json.NewDecoder(r)

	data := make(map[string]interface{})
	err := dec.Decode(&data)
	if err != nil {
		return "", fmt.Errorf("failed parsing json keyfile: %w", err)
	}

	tmp, ok := data["project_id"]
	if !ok {
		return "", fmt.Errorf("failed finding project_id in keyfile")
	}

	id, ok := tmp.(string)
	if !ok {
		return "", fmt.Errorf("project_id in keyfile is not a string")
	}
	return id, nil
}
