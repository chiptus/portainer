package exec

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/stretchr/testify/assert"
)

func Test_createEnvFile(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name         string
		stack        *portaineree.Stack
		expected     string
		expectedFile bool
	}{
		{
			name: "should not add env file option if stack doesn't have env variables",
			stack: &portaineree.Stack{
				ProjectPath: dir,
			},
			expected: "",
		},
		{
			name: "should not add env file option if stack's env variables are empty",
			stack: &portaineree.Stack{
				ProjectPath: dir,
				Env:         []portaineree.Pair{},
			},
			expected: "",
		},
		{
			name: "should add env file option if stack has env variables",
			stack: &portaineree.Stack{
				ProjectPath: dir,
				Env: []portaineree.Pair{
					{Name: "var1", Value: "value1"},
					{Name: "var2", Value: "value2"},
				},
			},
			expected: "var1=value1\nvar2=value2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := createEnvFile(tt.stack)

			if tt.expected != "" {
				assert.Equal(t, "stack.env", result)

				f, _ := os.Open(path.Join(dir, "stack.env"))
				content, _ := ioutil.ReadAll(f)

				assert.Equal(t, tt.expected, string(content))
			} else {
				assert.Equal(t, "", result)
			}
		})
	}
}

func Test_createNetworkEnvFile(t *testing.T) {
	dir := t.TempDir()
	buf := []byte(`
version: '3.6'
services:
  nginx-example:
    image: nginx:latest
networks:
  default:
    name: ${test}
    driver: bridge
`)
	if err := ioutil.WriteFile(path.Join(dir,
		"docker-compose.yml"), buf, 0644); err != nil {
		t.Fatalf("Failed to create yaml file: %s", err)
	}

	stackWithoutEnv := &portaineree.Stack{
		ProjectPath: dir,
		EntryPoint:  "docker-compose.yml",
		Env:         []portaineree.Pair{},
	}

	if err := createNetworkEnvFile(stackWithoutEnv); err != nil {
		t.Fatalf("Failed to create network env file: %s", err)
	}

	content, err := ioutil.ReadFile(path.Join(dir, ".env"))
	if err != nil {
		t.Fatalf("Failed to read network env file: %s", err)
	}

	assert.Equal(t, "test=None\n", string(content))

	stackWithEnv := &portaineree.Stack{
		ProjectPath: dir,
		EntryPoint:  "docker-compose.yml",
		Env: []portaineree.Pair{
			{Name: "test", Value: "test-value"},
		},
	}

	if err := createNetworkEnvFile(stackWithEnv); err != nil {
		t.Fatalf("Failed to create network env file: %s", err)
	}

	content, err = ioutil.ReadFile(path.Join(dir, ".env"))
	if err != nil {
		t.Fatalf("Failed to read network env file: %s", err)
	}

	assert.Equal(t, "test=test-value\n", string(content))
}
