package exec

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/joho/godotenv/autoload"

	portaineree "github.com/portainer/portainer-ee/api"
)

const composeFile = `version: "3.9"
services:
  busybox:
    image: "alpine:latest"
    container_name: "compose_wrapper_test"`
const composedContainerName = "compose_wrapper_test"

func setup(t *testing.T) (*portaineree.Stack, *portaineree.Endpoint) {
	dir := t.TempDir()
	composeFileName := "compose_wrapper_test.yml"
	f, _ := os.Create(filepath.Join(dir, composeFileName))
	f.WriteString(composeFile)

	stack := &portaineree.Stack{
		ProjectPath: dir,
		EntryPoint:  composeFileName,
		Name:        "project-name",
	}

	endpoint := &portaineree.Endpoint{
		URL: "unix://",
	}

	return stack, endpoint
}

func Test_UpAndDown(t *testing.T) {

	if !integrationTest() {
		return
	}

	stack, endpoint := setup(t)

	w, err := NewComposeStackManager("", "", nil)
	if err != nil {
		t.Fatalf("Failed creating manager: %s", err)
	}

	ctx := context.TODO()

	err = w.Up(ctx, stack, endpoint, false)
	if err != nil {
		t.Fatalf("Error calling docker-compose up: %s", err)
	}

	if !containerExists(composedContainerName) {
		t.Fatal("container should exist")
	}

	err = w.Down(ctx, stack, endpoint)
	if err != nil {
		t.Fatalf("Error calling docker-compose down: %s", err)
	}

	if containerExists(composedContainerName) {
		t.Fatal("container should be removed")
	}
}

func containerExists(containerName string) bool {
	cmd := exec.Command("docker", "ps", "-a", "-f", fmt.Sprintf("name=%s", containerName))

	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to list containers: %s", err)
	}

	return strings.Contains(string(out), containerName)
}

func integrationTest() bool {
	if val, ok := os.LookupEnv("INTEGRATION_TEST"); ok {
		return strings.EqualFold(val, "true")
	}

	return false
}
