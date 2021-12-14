package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	requesthelpers "github.com/portainer/libhttp/request"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
)

const (
	contextEndpoint = "endpoint"
)

func WithEndpoint(endpointService portainer.EndpointService, endpointIDParam string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, request *http.Request) {
			if endpointIDParam == "" {
				endpointIDParam = "id"
			}

			endpointID, err := requesthelpers.RetrieveNumericRouteVariableValue(request, endpointIDParam)
			if err != nil {
				httperror.WriteError(rw, http.StatusBadRequest, "Invalid environment identifier route variable", err)
				return
			}

			endpoint, err := endpointService.Endpoint(portainer.EndpointID(endpointID))
			if err != nil {
				statusCode := http.StatusInternalServerError

				if err == bolterrors.ErrObjectNotFound {
					statusCode = http.StatusNotFound
				}
				httperror.WriteError(rw, statusCode, "Unable to find an environment with the specified identifier inside the database", err)
				return
			}

			ctx := context.WithValue(request.Context(), contextEndpoint, endpoint)

			next.ServeHTTP(rw, request.WithContext(ctx))

		})
	}
}

func SetEndpoint(endpoint *portainer.Endpoint, request *http.Request) {
	ctx := context.WithValue(request.Context(), contextEndpoint, endpoint)
	*request = *request.WithContext(ctx)
}

type ContextFetcher func(request *http.Request) (*portainer.Endpoint, error)

func StaticEndpoint(endpoint *portainer.Endpoint) ContextFetcher {
	return func(request *http.Request) (*portainer.Endpoint, error) {
		return endpoint, nil
	}
}

func FetchEndpoint(request *http.Request) (*portainer.Endpoint, error) {
	contextData := request.Context().Value(contextEndpoint)
	if contextData == nil {
		return nil, errors.New("Unable to find environment data in request context")
	}

	return contextData.(*portainer.Endpoint), nil
}

// FindInQuery returns a func that finds a query param by name and returns a corresponding Endpoint.
// If either param or endpoint are missing, it returns an error
func FindInQuery(endpointService portainer.EndpointService, param string) ContextFetcher {
	return findInRequest(endpointService, getIntQueryParam(param))
}

// FindInPath returns a func that finds a url param by name and returns a corresponding Endpoint.
// If either param or endpoint are missing, it returns an error
func FindInPath(endpointService portainer.EndpointService, param string) ContextFetcher {
	return findInRequest(endpointService, getIntRouteParam(param))
}

// FindInJsonBodyField returns a func that finds a field by its path and returns a corresponding Endpoint.
// If request body is missing a requested field or endpoint is missing, it returns an error.
// FieldPath should represent a field hierarchy with the field holding the endpoint id being the last.
func FindInJsonBodyField(endpointService portainer.EndpointService, fieldPath []string) ContextFetcher {
	return findInRequest(endpointService, getIntJsonBodyField(fieldPath))
}

// findInRequest returns a func that looksup an endpoint Id in the request and returns a corresponding Endpoint.
// If either param or endpoint are missing, it returns an error
func findInRequest(endpointService portainer.EndpointService, lookup endpointIdLookup) ContextFetcher {
	return func(request *http.Request) (*portainer.Endpoint, error) {
		endpointID, err := lookup(request)
		if err != nil {
			return nil, err
		}

		endpoint, err := endpointService.Endpoint(endpointID)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't find an endpoint")
		}

		return endpoint, nil
	}
}

type endpointIdLookup func(*http.Request) (portainer.EndpointID, error)

func asPortainerID(v string) (portainer.EndpointID, error) {
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}

	return portainer.EndpointID(i), nil
}

func getIntQueryParam(param string) endpointIdLookup {
	return func(r *http.Request) (portainer.EndpointID, error) {
		queryParameter := r.FormValue(param)
		if queryParameter == "" {
			return 0, errors.Errorf("cannot find a query param %s", param)
		}

		return asPortainerID(queryParameter)
	}
}

func getIntRouteParam(param string) endpointIdLookup {
	return func(r *http.Request) (portainer.EndpointID, error) {
		routeVariables := mux.Vars(r)
		if routeVariables != nil {
			if routeVar, ok := routeVariables[param]; ok {
				return asPortainerID(routeVar)
			}
		}

		return 0, errors.Errorf("cannot find route param %s", param)
	}
}

func getIntJsonBodyField(fieldPath []string) endpointIdLookup {
	return func(r *http.Request) (portainer.EndpointID, error) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return 0, errors.Wrap(err, "cannot read request body")
		}

		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		var b map[string]interface{}
		if err := json.Unmarshal(body, &b); err != nil {
			return 0, errors.Wrap(err, "failed to unmarshal request payload")
		}

		for i, part := range fieldPath {
			val, ok := b[part]
			if !ok {
				return 0, errors.Wrapf(err, "failed to find specified path in the request payload: %s", part)
			}

			if i == len(fieldPath)-1 {
				// by default all digit-based values are converted to float64 by the unmarshalling,
				// we'd be treating it as a correct type and convert to int upon return
				value, ok := val.(float64)
				if !ok {
					return 0, errors.Errorf("body part %s doesn't seem to hold an id", part)
				}
				return portainer.EndpointID(value), nil
			}

			b, ok = val.(map[string]interface{})
			if !ok {
				return 0, errors.Errorf("body part %s is missing necessary nested fields", part)
			}
		}

		return 0, errors.New("couldn't find a requested body field")
	}
}
