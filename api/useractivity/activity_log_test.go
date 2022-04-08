package useractivity

import (
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/stretchr/testify/assert"
)

func createActivityLog(username, context, action string, payload []byte) *portaineree.UserActivityLog {
	return &portaineree.UserActivityLog{
		UserActivityLogBase: portaineree.UserActivityLogBase{
			Timestamp: time.Now().Unix(),
			Username:  username,
		},
		Context: context,
		Action:  action,
		Payload: payload,
	}
}

func TestAddUserActivity(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	expectedPayloadString := "payload"

	userLog := createActivityLog("username", "context", "action", []byte(expectedPayloadString))

	err = store.StoreUserActivityLog(userLog)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	var logs []*portaineree.UserActivityLog

	err = store.db.All(&logs)
	if err != nil {
		t.Fatalf("Failed retrieving activities: %s", err)
	}

	assert.Equal(t, 1, len(logs), "Store should have one element")
	assert.Equal(t, userLog, logs[0], "logs should be equal")
}

func TestGetUserActivityLogs(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createActivityLog("username1", "context1", "action1", []byte("payload1"))
	err = store.StoreUserActivityLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createActivityLog("username2", "context2", "action2", []byte("payload2"))
	err = store.StoreUserActivityLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createActivityLog("username3", "context3", "action3", []byte("payload3"))
	err = store.StoreUserActivityLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	logs, _, err := store.GetUserActivityLogs(portaineree.UserActivityLogBaseQuery{})
	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.UserActivityLog{log1, log2, log3}, logs)
}

func TestGetUserActivityLogsByTimestamp(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createActivityLog("username1", "context1", "action1", []byte("payload1"))
	err = store.StoreUserActivityLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	time.Sleep(time.Second * 1)

	log2 := createActivityLog("username2", "context2", "action2", []byte("payload2"))
	err = store.StoreUserActivityLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	time.Sleep(time.Second * 1)

	log3 := createActivityLog("username3", "context3", "action3", []byte("payload3"))
	err = store.StoreUserActivityLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	logs, _, err := store.GetUserActivityLogs(portaineree.UserActivityLogBaseQuery{
		BeforeTimestamp: log3.Timestamp - 1,
		AfterTimestamp:  log1.Timestamp + 1,
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, log2, logs[0], "logs are not equal")
}

func TestGetUserActivityLogsByKeyword(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createActivityLog("username1", "context1", "action1", []byte("success"))
	err = store.StoreUserActivityLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createActivityLog("username2", "context2", "action2", []byte("error"))
	err = store.StoreUserActivityLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createActivityLog("username3", "context3", "action3", []byte("success"))
	err = store.StoreUserActivityLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	// like
	shouldHaveAllLogs, _, err := store.GetUserActivityLogs(portaineree.UserActivityLogBaseQuery{
		Keyword: "username",
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.UserActivityLog{log1, log2, log3}, shouldHaveAllLogs)

	// username
	shouldHaveOnlyLog1, _, err := store.GetUserActivityLogs(portaineree.UserActivityLogBaseQuery{
		Keyword: "username1",
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, log1, shouldHaveOnlyLog1[0])

	// action
	shouldHaveOnlyLog3, _, err := store.GetUserActivityLogs(portaineree.UserActivityLogBaseQuery{
		Keyword: "action3",
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, 1, len(shouldHaveOnlyLog3))
	assert.Equal(t, log3, shouldHaveOnlyLog3[0])
}

func TestGetUserActivityLogsSortOrderAndPaginate(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createActivityLog("username1", "context1", "action1", []byte("payload1"))
	err = store.StoreUserActivityLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createActivityLog("username2", "context2", "action2", []byte("payload2"))
	err = store.StoreUserActivityLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createActivityLog("username3", "context3", "action3", []byte("payload3"))
	err = store.StoreUserActivityLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log4 := createActivityLog("username4", "context4", "action4", []byte("payload4"))
	err = store.StoreUserActivityLog(log4)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	shouldBeLog4AndLog3, _, err := store.GetUserActivityLogs(portaineree.UserActivityLogBaseQuery{
		SortDesc: true,
		SortBy:   "Username",
		Offset:   0,
		Limit:    2,
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.UserActivityLog{log4, log3}, shouldBeLog4AndLog3)

	shouldBeLog2AndLog1, _, err := store.GetUserActivityLogs(portaineree.UserActivityLogBaseQuery{
		SortDesc: true,
		SortBy:   "Username",
		Offset:   2,
		Limit:    2,
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.UserActivityLog{log2, log1}, shouldBeLog2AndLog1)
}

func TestGetUserActivityLogsDesc(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createActivityLog("username1", "context1", "action1", []byte("payload1"))
	err = store.StoreUserActivityLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createActivityLog("username2", "context2", "action2", []byte("payload2"))
	err = store.StoreUserActivityLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createActivityLog("username3", "context3", "action3", []byte("payload3"))
	err = store.StoreUserActivityLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	logs, _, err := store.GetUserActivityLogs(portaineree.UserActivityLogBaseQuery{
		SortDesc: true,
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.UserActivityLog{log3, log2, log1}, logs)
}

func TestDoubleClose(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	store.Close()
	store.Close()
}
