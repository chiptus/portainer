package endpoints

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"
	bolt "github.com/portainer/portainer/api/bolt/bolttest"
	"github.com/portainer/portainer/api/http/security"
	helper "github.com/portainer/portainer/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_endpointUpdate(t *testing.T) {
	is := assert.New(t)

	store, teardown := bolt.MustNewTestStore(true)
	defer teardown()

	err := store.Endpoint().CreateEndpoint(&portainer.Endpoint{ID: 1})
	is.NoError(err, "error creating environment")

	err = store.User().CreateUser(&portainer.User{Username: "admin", Role: portainer.AdministratorRole})
	is.NoError(err, "error creating a user")

	bouncer := helper.NewTestRequestBouncer()
	h := NewHandler(bouncer)
	h.DataStore = store
	h.UserActivityStore = helper.NewUserActivityStore()

	t.Run("Test valid autoUpdate settings", func(t *testing.T) {
		start := "00:00"
		end := "23:59"

		endpointSettings := portainer.Endpoint{
			ChangeWindow: portainer.EndpointChangeWindow{
				Enabled:   true,
				StartTime: start,
				EndTime:   end,
			},
		}

		data, err := json.Marshal(endpointSettings)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPut, "/endpoints/1/settings", bytes.NewBuffer(data))
		ctx := security.StoreTokenData(req, &portainer.TokenData{ID: 1, Username: "admin", Role: 1})
		req = req.WithContext(ctx)
		req.Header.Add("Authorization", "Bearer dummytoken")

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code, "Status should be 200")

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		resp := portainer.Endpoint{}
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be json")
		is.EqualValues(true, resp.ChangeWindow.Enabled, "Enabled doesn't match")
		is.EqualValues(start, resp.ChangeWindow.StartTime, "StartTime doesn't match")
		is.EqualValues(end, resp.ChangeWindow.EndTime, "EndTime doesn't match")
	})

	t.Run("Test invalid autoUpdate time range", func(t *testing.T) {
		endpointSettings := portainer.Endpoint{
			ChangeWindow: portainer.EndpointChangeWindow{
				Enabled:   true,
				StartTime: "99:00",
				EndTime:   "23:59",
			},
		}

		data, err := json.Marshal(endpointSettings)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPut, "/endpoints/1/settings", bytes.NewBuffer(data))
		ctx := security.StoreTokenData(req, &portainer.TokenData{ID: 1, Username: "admin", Role: 1})
		req = req.WithContext(ctx)
		req.Header.Add("Authorization", "Bearer dummytoken")

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusBadRequest, rr.Code, "Status should be 400")
	})
}
