package auth

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/logoutcontext"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/rs/zerolog/log"
)

// @id Logout
// @summary Logout
// @description **Access policy**: public
// @security ApiKeyAuth
// @security jwt
// @tags auth
// @success 204 "Success"
// @failure 500 "Server error"
// @router /auth/logout [post]
func (handler *Handler) logout(w http.ResponseWriter, r *http.Request) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Username: "Unauthenticated user",
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		log.Warn().Err(err).Msg("unable to retrieve user details from authentication token")
		return resp, response.Empty(w)
	}

	if tokenData != nil {
		resp.Username = tokenData.Username
		handler.KubernetesTokenCacheManager.RemoveUserFromCache(tokenData.ID)
		logoutcontext.Cancel(tokenData.Token)
	}

	return resp, response.Empty(w)
}
