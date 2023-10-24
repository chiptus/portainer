package motd

import (
	"net/http"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/client"
	"github.com/portainer/portainer/pkg/libcrypto"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/segmentio/encoding/json"
)

type motdResponse struct {
	Title         string            `json:"Title"`
	Message       string            `json:"Message"`
	ContentLayout map[string]string `json:"ContentLayout"`
	Style         string            `json:"Style"`
	Hash          []byte            `json:"Hash"`
}

type motdData struct {
	Title         string            `json:"title"`
	Message       []string          `json:"message"`
	ContentLayout map[string]string `json:"contentLayout"`
	Style         string            `json:"style"`
}

// @id MOTD
// @summary fetches the message of the day
// @description **Access policy**: restricted
// @tags motd
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {object} motdResponse
// @router /motd [get]
func (handler *Handler) motd(w http.ResponseWriter, r *http.Request) {
	motd, err := client.Get(portaineree.MessageOfTheDayURL, 0)
	if err != nil {
		response.JSON(w, &motdResponse{Message: ""})
		return
	}

	var data motdData
	err = json.Unmarshal(motd, &data)
	if err != nil {
		response.JSON(w, &motdResponse{Message: ""})
		return
	}

	message := strings.Join(data.Message, "\n")

	hash := libcrypto.HashFromBytes([]byte(message))
	resp := motdResponse{
		Title:         data.Title,
		Message:       message,
		Hash:          hash,
		ContentLayout: data.ContentLayout,
		Style:         data.Style,
	}

	response.JSON(w, &resp)
}
