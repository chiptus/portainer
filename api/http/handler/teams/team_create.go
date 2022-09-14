package teams

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

type teamCreatePayload struct {
	// Name
	Name string `example:"developers" validate:"required"`
	// TeamLeaders
	TeamLeaders []portaineree.UserID `example:"3,5"`
}

func (payload *teamCreatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid team name")
	}
	return nil
}

// @id TeamCreate
// @summary Create a new team
// @description Create a new team.
// @description **Access policy**: administrator
// @tags teams
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body teamCreatePayload true "details"
// @success 200 {object} portaineree.Team "Success"
// @failure 400 "Invalid request"
// @failure 409 "Team already exists"
// @failure 500 "Server error"
// @router /team [post]
func (handler *Handler) teamCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload teamCreatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	team, err := handler.DataStore.Team().TeamByName(payload.Name)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return httperror.InternalServerError("Unable to retrieve teams from the database", err)
	}
	if team != nil {
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: "A team with the same name already exists", Err: errors.New("Team already exists")}
	}

	team = &portaineree.Team{
		Name: payload.Name,
	}

	err = handler.DataStore.Team().Create(team)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the team inside the database", err)
	}

	for _, teamLeader := range payload.TeamLeaders {
		membership := &portaineree.TeamMembership{
			UserID: teamLeader,
			TeamID: team.ID,
			Role:   portaineree.TeamLeader,
		}

		err = handler.DataStore.TeamMembership().Create(membership)
		if err != nil {
			return httperror.InternalServerError("Unable to persist team leadership inside the database", err)
		}
	}

	return response.JSON(w, team)
}
