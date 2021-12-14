package integration_tests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	host     string
	username string
	password string
)

func TestMain(m *testing.M) {
	h, hostOk := os.LookupEnv("API_LOGGING_HOST")
	u, usernameOk := os.LookupEnv("API_LOGGING_USERNAME")
	p, passwordOk := os.LookupEnv("API_LOGGING_PASSWORD")
	if !hostOk || !usernameOk || !passwordOk {
		fmt.Println(`This integration tests need a running standalone environment.
		Please set following env vars: API_LOGGING_HOST, API_LOGGING_USERNAME, API_LOGGING_PASSWORD.`)
		return
	}

	host, username, password = h, u, p

	// random seen need to be setup to produce rand values on each run, otherwise values would be the same across runs
	rand.Seed(time.Now().UnixNano())

	m.Run()
}

func Test_Template_Create_FromString(t *testing.T) {
	token := login()

	templateId := strconv.Itoa(rand.Int())
	payload := `{"Title":"template` + templateId + `","FileContent":"version: \"3\"\n\nservices:\n  busybox:\n    image: busybox\n","File":null,"RepositoryURL":"","RepositoryReferenceName":"","RepositoryAuthentication":false,"RepositoryUsername":"","RepositoryPassword":"","ComposeFilePathInRepository":"docker-compose.yml","Description":"description","Note":"","Logo":"","Platform":1,"Type":1,"AccessControlData":{"AccessControlEnabled":true,"Ownership":"administrators","AuthorizedUsers":[],"AuthorizedTeams":[]}}`

	if _, ok := post(token, `/custom_templates?method=string`, []byte(payload)); !ok {
		t.Fail()
	}

	entry := lastLogEntry(token)
	require.NotNil(t, entry)
	assert.Equal(t, entry.Username, username)
	assert.Equal(t, entry.Context, "Portainer")

	sanitizedPayload := `{"Title":"template` + templateId + `","FileContent":"version: \"3\"\n\nservices:\n  busybox:\n    image: busybox\n","File":null,"RepositoryURL":"","RepositoryReferenceName":"","RepositoryAuthentication":false,"RepositoryUsername":"","RepositoryPassword":"[REDACTED]","ComposeFilePathInRepository":"docker-compose.yml","Description":"description","Note":"","Logo":"","Platform":1,"Type":1,"AccessControlData":{"AccessControlEnabled":true,"Ownership":"administrators","AuthorizedUsers":[],"AuthorizedTeams":[]}}`
	assert.JSONEq(t, entry.Payload, sanitizedPayload, `Payload should be the same as the one submitted, sensitive fields being sanitized as [REDACTED]`)
}

func Test_Template_Create_FromFile(t *testing.T) {
	token := login()

	tmpdir := t.TempDir()
	filename := path.Join(tmpdir, "busybox.yml")
	os.WriteFile(filename, []byte("version: \"3\"\n\nservices:\n  busybox:\n    image: busybox\n"), 0644)

	templateId := strconv.Itoa(rand.Int())
	payload := `--xxx
Content-Disposition: form-data; name="Title"

template` + templateId + `
--xxx
Content-Disposition: form-data; name="Description"

file
--xxx
Content-Disposition: form-data; name="Platform"

1
--xxx
Content-Disposition: form-data; name="Type"

1
--xxx
Content-Disposition: form-data; name="AccessControlData[AccessControlEnabled]"

true
--xxx
Content-Disposition: form-data; name="AccessControlData[Ownership]"

administrators
--xxx
Content-Disposition: form-data; name="File"; filename="` + filename + `"
Content-Type: application/x-yaml


--xxx--
`

	if _, ok := post(token, `/custom_templates?method=file`, []byte(payload), header{`Content-Type`, `multipart/form-data; boundary=xxx`}); !ok {
		t.Fail()
	}

	entry := lastLogEntry(token)
	require.NotNil(t, entry)
	assert.Equal(t, entry.Username, username)
	assert.Equal(t, entry.Context, "Portainer")

	sanitizedPayload := `{"Title":"template` + templateId + `", "Description":"file", "Platform":"1", "Type":"1", "method":"file", "AccessControlData[AccessControlEnabled]":"true", "AccessControlData[Ownership]":"administrators"}`
	assert.JSONEq(t, entry.Payload, sanitizedPayload, `Payload should be the same as the one submitted, file should be skipped`)
}

func Test_Stack_Create_Standalone(t *testing.T) {
	token := login()

	id := strconv.Itoa(rand.Int())
	payload := `{"Name":"busybox` + id + `", "StackFileContent":"version: \"3\"\n\nservices:\n  busybox:\n    image: busybox\n","Env":[]}`

	createStackUrl := `/stacks?endpointId=1&method=string&type=2`
	if _, ok := post(token, createStackUrl, []byte(payload)); !ok {
		t.Fail()
	}

	entry := lastLogEntry(token)
	require.NotNil(t, entry)
	assert.Equal(t, entry.Username, username)
	assert.Equal(t, entry.Context, "local")
	assert.Equal(t, entry.Action, `POST `+createStackUrl)
	assert.JSONEq(t, entry.Payload, payload, `Payload should be the same as the one submitted, sensitive fields being sanitized as [REDACTED]`)
}

func Test_Stack_Start_Stop_Standalone(t *testing.T) {
	token := login()

	payload := `{"Name":"busybox` + strconv.Itoa(rand.Int()) + `", "StackFileContent":"version: \"3\"\n\nservices:\n  busybox:\n    image: busybox\n","Env":[]}`

	reply, ok := post(token, `/stacks?endpointId=1&method=string&type=2`, []byte(payload))
	if !ok {
		t.Fail()
	}

	var createResponse struct{ Id int }
	json.Unmarshal(reply, &createResponse)

	stopStackUrl := fmt.Sprintf("/stacks/%d/stop?endpointId=1", createResponse.Id)
	if _, ok := post(token, stopStackUrl, nil); !ok {
		t.Fatal()
	}

	entry := lastLogEntry(token)
	require.NotNil(t, entry)
	assert.Equal(t, entry.Username, username)
	assert.Equal(t, entry.Context, "local")
	assert.Equal(t, entry.Action, `POST `+stopStackUrl)
	assert.Equal(t, entry.Payload, ``)

	startStackUrl := fmt.Sprintf("/stacks/%d/start", createResponse.Id)
	if _, ok := post(token, startStackUrl, nil); !ok {
		t.Fail()
	}

	entry = lastLogEntry(token)
	require.NotNil(t, entry)
	assert.Equal(t, entry.Username, username)
	assert.Equal(t, entry.Context, "local")
	assert.Equal(t, entry.Action, `POST `+startStackUrl)
	assert.Equal(t, entry.Payload, ``)
}

//
// Helpers
//

func login() (token string) {
	r, _ := http.NewRequest(http.MethodPost, host+`/api/auth`, strings.NewReader(fmt.Sprintf(`{"Username":"%s", "Password":"%s"}`, username, password)))

	response, err := http.DefaultClient.Do(r)
	if err != nil {
		return ``
	}

	v, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ``
	}
	response.Body.Close()

	if response.StatusCode != 200 {
		fmt.Printf("Login failed. Response: %s, Status: %v\n", v, response.StatusCode)
	}

	var body struct {
		Jwt string
	}

	if err := json.Unmarshal(v, &body); err != nil {
		return ``
	}

	return body.Jwt
}

type logEntry struct {
	Timestamp int
	Action    string
	Username  string
	Context   string
	Payload   string
}

func lastLogEntry(token string) *logEntry {
	r, _ := http.NewRequest(http.MethodGet, host+`/api/useractivity/logs`, nil)
	r.Header.Add("Authorization", `Bearer `+token)

	response, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil
	}

	v, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil
	}
	response.Body.Close()

	if response.StatusCode != 200 {
		fmt.Printf("Log fetch failed. Response: %s, Status: %v\n", v, response.StatusCode)
	}

	var entries struct {
		Logs []logEntry
	}
	if err := json.Unmarshal(v, &entries); err != nil {
		return nil
	}

	var lastEntry logEntry
	for _, e := range entries.Logs {
		if e.Timestamp > lastEntry.Timestamp {
			lastEntry = e
		}
	}

	payload, _ := base64.StdEncoding.DecodeString(lastEntry.Payload)
	lastEntry.Payload = string(payload)

	return &lastEntry
}

type header struct {
	key   string
	value string
}

func post(token string, path string, payload []byte, headers ...header) (reply []byte, ok bool) {
	r, _ := http.NewRequest(http.MethodPost, host+`/api`+path, bytes.NewReader(payload))
	r.Header.Add("Authorization", `Bearer `+token)

	if payload != nil {
		r.Header.Add("Content-Type", "application/json")
	}

	for _, h := range headers {
		r.Header.Set(h.key, h.value)
	}

	response, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, false
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, false
	}
	response.Body.Close()

	if response.StatusCode != 200 {
		fmt.Printf("Request failed. Response: %s, Status: %v\n", body, response.StatusCode)
		return nil, false
	}

	return body, true
}
