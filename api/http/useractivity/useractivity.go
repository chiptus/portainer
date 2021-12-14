package useractivity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/middlewares"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/utils"
	"github.com/portainer/portainer/api/useractivity"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

// LogUserActivity a user activity logging middleware
// It relies on the endpoint being supplies throug the middleware.WithEndpoint.
// The endpoint will be used as a logging context, alternatively Portainer would be used as a context
func LogUserActivity(service portainer.UserActivityService) func(http.Handler) http.Handler {
	return LogUserActivityWithContext(service, middlewares.FetchEndpoint)
}

// LogUserActivityWithContext a user activity logging middleware
// It relies on the middlewares.ContextFetcher to fetch a logging context (i.e. endpoint).
// Alternatively Portainer would be used as a context
func LogUserActivityWithContext(service portainer.UserActivityService, context middlewares.ContextFetcher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// need a copy of the request because after calling next.ServeHTTP(w, r) request body will be empty and closed
			body := utils.CopyRequestBody(r)

			// overriding response writer with a custom one to have access to the written status code
			sw := negroni.NewResponseWriter(w)
			next.ServeHTTP(sw, r)

			if isGoodToLog(r.Method, sw.Status()) {
				LogActivity(service, context, body, r)
			}
		})
	}
}

// LogProxiedActivity logs a user activity for proxied requests
// It relies on the middlewares.ContextFetcher to fetch a logging context (i.e. endpoint).
// Alternatively Portainer would be used as a context.
// requestStatus represents the http status code of the proxied request.
func LogProxiedActivity(service portainer.UserActivityService, endpoint *portainer.Endpoint, responseStatus int, body []byte, r *http.Request) {
	if isGoodToLog(r.Method, responseStatus) {
		LogActivity(service, middlewares.StaticEndpoint(endpoint), body, r)
	}
}

// a check to define if a given http call should be logged or not
func isGoodToLog(requestMethod string, responseStatus int) bool {
	isModifyRequest := requestMethod == "POST" || requestMethod == "DELETE" || requestMethod == "PUT" || requestMethod == "PATCH"
	requestSucceeded := 200 <= responseStatus && responseStatus < 300
	return isModifyRequest && requestSucceeded
}

func LogActivity(service portainer.UserActivityService, contextFetcher middlewares.ContextFetcher, body []byte, r *http.Request) {
	var err error

	contentType := r.Header.Get("Content-Type")
	switch strings.Split(contentType, ";")[0] {
	case "multipart/form-data", "application/x-www-form-urlencoded":
		const defaultMaxMemory = 32 << 20 // 32 MB
		r.ParseMultipartForm(defaultMaxMemory)

		// only capture santized form values and skip files
		b := make(map[string]interface{})
		for k, v := range r.Form {
			if len(v) == 1 {
				b[k] = v[0]
			} else {
				b[k] = v
			}
		}

		b = useractivity.Sanitise(b)

		body, err = json.Marshal(b)
		if err != nil {
			logrus.WithError(err).Debug("failed to marshal user activity payload")
			return
		}

	case "application/json":
		var b map[string]interface{}
		if err := json.Unmarshal(body, &b); err != nil {
			logrus.WithError(err).Debug("failed to unmarshal user activity payload")
			return
		}

		b = useractivity.Sanitise(b)

		body, err = json.Marshal(b)
		if err != nil {
			logrus.WithError(err).Debug("failed to marshal user activity payload")
			return
		}

	default:
		// ignore the other body types assuming they are either files or unimportant
		body = nil
	}

	username := ""
	tokenData, err := security.RetrieveTokenData(r)
	if err == nil && tokenData != nil {
		username = tokenData.Username
	}

	action := fmt.Sprintf("%s %s", r.Method, r.URL.String())

	context := "Portainer"
	endpoint, err := contextFetcher(r)
	if err == nil && endpoint != nil {
		context = endpoint.Name
	}

	err = service.LogUserActivity(username, context, action, body)
	if err != nil {
		logrus.WithError(err).Debug("failed logging user activity")
	}
}
