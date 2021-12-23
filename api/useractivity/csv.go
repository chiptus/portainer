package useractivity

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
)

// MarshalAuthLogsToCSV converts a list of logs to a CSV string
func MarshalAuthLogsToCSV(w io.Writer, logs []*portaineree.AuthActivityLog) error {
	var headers = []string{
		"Time",
		"Origin",
		"Context",
		"Username",
		"Result",
	}

	csvw := csv.NewWriter(w)

	err := csvw.Write(headers)
	if err != nil {
		return err
	}

	for _, log := range logs {
		result := ""
		switch log.Type {
		case portaineree.AuthenticationActivityFailure:
			result = "Authentication failure"
		case portaineree.AuthenticationActivitySuccess:
			result = "Authentication success"
		case portaineree.AuthenticationActivityLogOut:
			result = "Logout"
		}

		context := ""
		switch log.Context {
		case portaineree.AuthenticationInternal:
			context = "Internal"
		case portaineree.AuthenticationLDAP:
			context = "LDAP"
		case portaineree.AuthenticationOAuth:
			context = "OAuth"
		}

		timestamp := time.Unix(log.Timestamp, 0)
		formattedTimestamp := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			timestamp.Year(), timestamp.Month(), timestamp.Day(),
			timestamp.Hour(), timestamp.Minute(), timestamp.Second())

		err := csvw.Write([]string{
			formattedTimestamp,
			log.Origin,
			context,
			log.Username,
			result,
		})
		if err != nil {
			return err
		}

	}
	csvw.Flush()

	return csvw.Error()
}

// MarshalLogsToCSV converts a list of logs to a CSV string
func MarshalLogsToCSV(w io.Writer, logs []*portaineree.UserActivityLog) error {
	var headers = []string{
		"Time",
		"Username",
		"Environment",
		"Action",
		"Payload",
	}

	csvw := csv.NewWriter(w)

	err := csvw.Write(headers)
	if err != nil {
		return err
	}

	for _, log := range logs {

		timestamp := time.Unix(log.Timestamp, 0)
		formattedTimestamp := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			timestamp.Year(), timestamp.Month(), timestamp.Day(),
			timestamp.Hour(), timestamp.Minute(), timestamp.Second())

		err := csvw.Write([]string{
			formattedTimestamp,
			log.Username,
			log.Context,
			log.Action,
			string(log.Payload),
		})
		if err != nil {
			return err
		}
	}

	csvw.Flush()

	return csvw.Error()
}
