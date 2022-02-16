package errors

import "errors"

var (
	// TODO: i'm pretty sure this needs wrapping at several levels
	ErrWrongDBEdition = errors.New("the Portainer database is set for Portainer Business Edition, please follow the instructions in our documentation to downgrade it: https://documentation.portainer.io/v2.0-be/downgrade/be-to-ce/")
	ErrMigrationToCE  = errors.New("DB is already on CE edition")
)
