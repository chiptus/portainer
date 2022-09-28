package auth

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/http/security"
)

// @id Logout
// @summary Logout
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags auth
// @success 204 "Success"
// @failure 500 "Server error"
// @router /auth/logout [post]
func (handler *Handler) logout(w http.ResponseWriter, r *http.Request) (*authMiddlewareResponse, *httperror.HandlerError) {
	tokenData, err := security.RetrieveTokenData(r)
	resp := &authMiddlewareResponse{
		Username: tokenData.Username,
	}

	if err != nil {
		return resp, httperror.InternalServerError("Unable to retrieve user details from authentication token", err)
	}

	handler.KubernetesTokenCacheManager.RemoveUserFromCache(int(tokenData.ID))

	return resp, response.Empty(w)
}
