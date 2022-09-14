package useractivity

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/useractivity"
)

// @id AuthLogsCSV
// @summary Download auth logs as CSV
// @description Download auth logs as CSV by provided query
// @description **Access policy**: admin
// @tags useractivity
// @security ApiKeyAuth
// @security jwt
// @produce text/csv
// @param before query int false "Results before timestamp (unix)"
// @param after query int false "Results after timestamp (unix)"
// @param sortBy query string false "Sort by this column" Enum("Type", "Timestamp", "Origin", "Context", "Username", "Result")
// @param sortDesc query bool false "Sort order, if true will return results by descending order"
// @param limit query int false "Limit results"
// @param keyword query string false "Query logs by this keyword"
// @success 200 "Success"
// @failure 500 "Server error"
// @router /useractivity/authlogs.csv [get]
func (handler *Handler) authLogsCSV(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	limit, _ := request.RetrieveNumericQueryParameter(r, "limit", true)
	before, _ := request.RetrieveNumericQueryParameter(r, "before", true)
	after, _ := request.RetrieveNumericQueryParameter(r, "after", true)
	sortBy, _ := request.RetrieveQueryParameter(r, "sortBy", true)
	sortDesc, _ := request.RetrieveBooleanQueryParameter(r, "sortDesc", true)
	keyword, _ := request.RetrieveQueryParameter(r, "keyword", true)

	contextTypes, err := parseContextTypes(r.URL.RawQuery)
	if err != nil {
		return httperror.InternalServerError("Unable to parse query string", err)
	}

	activityTypes, err := parseActivityTypes(r.URL.RawQuery)
	if err != nil {
		return httperror.InternalServerError("Unable to parse query string", err)
	}

	opts := portaineree.AuthLogsQuery{
		UserActivityLogBaseQuery: portaineree.UserActivityLogBaseQuery{
			BeforeTimestamp: int64(before),
			AfterTimestamp:  int64(after),
			SortBy:          sortBy,
			SortDesc:        sortDesc,
			Keyword:         keyword,
			Limit:           limit,
		},
		ContextTypes:  contextTypes,
		ActivityTypes: activityTypes,
	}

	logs, _, err := handler.UserActivityStore.GetAuthLogs(opts)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve authentication logs", err)
	}

	err = useractivity.MarshalAuthLogsToCSV(w, logs)
	if err != nil {
		return httperror.InternalServerError("Unable to marshal logs to csv", err)
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"logs.csv\"")
	w.Header().Set("Content-Type", "text/csv")

	return nil
}
