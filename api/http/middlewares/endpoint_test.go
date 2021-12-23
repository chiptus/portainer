package middlewares

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	portaineree "github.com/portainer/portainer-ee/api"
	i "github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_FindInQuery(t *testing.T) {
	endpointService := i.NewDatastore(i.WithEndpoints([]portaineree.Endpoint{{ID: 1, Name: "EP"}})).Endpoint()

	cases := []struct {
		title    string
		url      string
		hasError bool
	}{
		{
			title:    "missing query params",
			url:      "/foo",
			hasError: true,
		},
		{
			title:    "missing endpoint in db",
			url:      "/foo?endpointId=2",
			hasError: true,
		},
		{
			title:    "missing requested query param",
			url:      "/foo?id=1",
			hasError: true,
		},
		{
			title:    "has requested param",
			url:      "/foo?endpointId=1",
			hasError: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, tt.url, nil)

			_, err := FindInQuery(endpointService, "endpointId")(r)
			assert.Equal(t, tt.hasError, err != nil)
		})
	}
}

func Test_FindInPath(t *testing.T) {
	endpointService := i.NewDatastore(i.WithEndpoints([]portaineree.Endpoint{{ID: 1, Name: "EP"}})).Endpoint()

	cases := []struct {
		title    string
		urlVars  map[string]string
		hasError bool
	}{
		{
			title:    "missing path vars",
			urlVars:  map[string]string{},
			hasError: true,
		},
		{
			title:    "missing endpoint in db",
			urlVars:  map[string]string{"endpointId": "2"},
			hasError: true,
		},
		{
			title:    "missing requested path param",
			urlVars:  map[string]string{"foo": "bar"},
			hasError: true,
		},
		{
			title:    "has requested path param",
			urlVars:  map[string]string{"endpointId": "1"},
			hasError: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			r = mux.SetURLVars(r, tt.urlVars)

			_, err := FindInPath(endpointService, "endpointId")(r)
			assert.Equal(t, tt.hasError, err != nil, "err: %s", err)
		})
	}
}

func Test_FindInJsonBodyField(t *testing.T) {
	endpointService := i.NewDatastore(i.WithEndpoints([]portaineree.Endpoint{{ID: 1, Name: "EP"}})).Endpoint()

	cases := []struct {
		title     string
		body      string
		fieldPath []string
		hasError  bool
	}{
		{
			title:     "shouldn't find in an empty body",
			body:      ``,
			fieldPath: []string{"top", "middle", "tail"},
			hasError:  true,
		},
		{
			title:     "shouldn't find when missing a top-level field",
			body:      `{"middle":{"tail":1}}`,
			fieldPath: []string{"top", "middle", "tail"},
			hasError:  true,
		},
		{
			title:     "shouldn't find when missing a middle-level field",
			body:      `{"top":{"tail":1}}`,
			fieldPath: []string{"top", "middle", "tail"},
			hasError:  true,
		},
		{
			title:     "shouldn't find when missing a tail-level field",
			body:      `{"top":{"middle":1}}`,
			fieldPath: []string{"top", "middle", "tail"},
			hasError:  true,
		},
		{
			title:     "shouldn't find when field has children",
			body:      `{"top":{"middle":{"tail":{"period":1}}}}`,
			fieldPath: []string{"top", "middle", "tail"},
			hasError:  true,
		},
		{
			title:     "shouldn't find when field has a different type",
			body:      `{"top":{"middle":{"tail":"1"}}}`,
			fieldPath: []string{"top", "middle", "tail"},
			hasError:  true,
		},
		{
			title:     "shouldn't find when corresponding endpoint is missing",
			body:      `{"top":{"middle":{"tail":2}}}`,
			fieldPath: []string{"top", "middle", "tail"},
			hasError:  true,
		},
		{
			title:     "should find when field when long path exists",
			body:      `{"top":{"middle":{"tail":1}}}`,
			fieldPath: []string{"top", "middle", "tail"},
			hasError:  false,
		},
		{
			title:     "should find when field when short path exists",
			body:      `{"tail":1, "foo":"bar"}`,
			fieldPath: []string{"tail"},
			hasError:  false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			body := bytes.NewBuffer([]byte(tt.body))
			r := httptest.NewRequest(http.MethodPost, "/", body)

			_, err := FindInJsonBodyField(endpointService, tt.fieldPath)(r)
			assert.Equal(t, tt.hasError, err != nil, "err: %s", err)

			b, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, []byte(tt.body), b, "request body shouldn't be changed by the find")
		})
	}
}
