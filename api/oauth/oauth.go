package oauth

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"

	portaineree "github.com/portainer/portainer-ee/api"
)

// Service represents a service used to authenticate users against an authorization server
type Service struct{}

// NewService returns a pointer to a new instance of this service
func NewService() *Service {
	return &Service{}
}

// Authenticate takes an access code and exchanges it for an access token from portainer OAuthSettings token environment(endpoint).
// On success, it will then return an OAuthInfo struct associated to authenticated user.
// The OAuthInfo struct contains data associated to the authenticated user.
// This data is obtained from the OAuth providers resource server and matched with the attributes of the user identifier(s).
func (*Service) Authenticate(code string, configuration *portaineree.OAuthSettings) (*portaineree.OAuthInfo, error) {
	token, err := getOAuthToken(code, configuration)
	if err != nil {
		log.Printf("[DEBUG] [internal,oauth] [message: failed retrieving oauth token: %v]", err)
		return nil, err
	}

	resource, err := getResource(token.AccessToken, configuration)
	if err != nil {
		log.Printf("[DEBUG] [internal,oauth] [message: failed retrieving resource: %v]", err)
		return nil, err
	}

	username, err := getUsername(resource, configuration)
	if err != nil {
		log.Printf("[DEBUG] [internal,oauth] [message: failed retrieving username: %v]", err)
		return nil, err
	}

	teams, err := getTeams(resource, configuration)
	if err != nil {
		log.Printf("[DEBUG] [internal,oauth] [message: failed retrieving oauth teams: %v]", err)
		return nil, err
	}

	return &portaineree.OAuthInfo{Username: username, Teams: teams}, nil
}

func getOAuthToken(code string, configuration *portaineree.OAuthSettings) (*oauth2.Token, error) {
	unescapedCode, err := url.QueryUnescape(code)
	if err != nil {
		return nil, err
	}

	config := buildConfig(configuration)
	token, err := config.Exchange(context.Background(), unescapedCode)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func getResource(token string, configuration *portaineree.OAuthSettings) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", configuration.ResourceURI, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &oauth2.RetrieveError{
			Response: resp,
			Body:     body,
		}
	}

	content, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	if content == "application/x-www-form-urlencoded" || content == "text/plain" {
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}

		datamap := make(map[string]interface{})
		for k, v := range values {
			if len(v) == 0 {
				datamap[k] = ""
			} else {
				datamap[k] = v[0]
			}
		}
		return datamap, nil
	}

	var datamap map[string]interface{}
	if err = json.Unmarshal(body, &datamap); err != nil {
		return nil, err
	}

	return datamap, nil
}

func buildConfig(configuration *portaineree.OAuthSettings) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  configuration.AuthorizationURI,
		TokenURL: configuration.AccessTokenURI,
	}

	return &oauth2.Config{
		ClientID:     configuration.ClientID,
		ClientSecret: configuration.ClientSecret,
		Endpoint:     endpoint,
		RedirectURL:  configuration.RedirectURI,
		Scopes:       []string{configuration.Scopes},
	}
}
