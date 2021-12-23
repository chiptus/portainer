package useractivity

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	portaineree "github.com/portainer/portainer-ee/api"
)

const (
	testType = portaineree.AuthenticationActivityType(0)
)

func createAuthLog(username, origin string, context portaineree.AuthenticationMethod, activityType portaineree.AuthenticationActivityType) *portaineree.AuthActivityLog {
	return &portaineree.AuthActivityLog{
		Type: activityType,
		UserActivityLogBase: portaineree.UserActivityLogBase{
			Timestamp: time.Now().Unix(),
			Username:  username,
		},
		Origin:  origin,
		Context: context,
	}
}

func BenchmarkAuthLog(b *testing.B) {
	defer timeTrack(time.Now(), "AuthActivityLog")

	// https://github.com/golang/go/issues/41062
	// bug in go 1.15 causes b.TempDir() to break in benchmarks
	// TODO remove in go 1.16

	err := os.RemoveAll("./useractivity.db")
	if err != nil {
		b.Fatalf("Failed removing db: %s", err)
	}

	store, err := setup("")
	if err != nil {
		b.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	authLog := createAuthLog("username", "endpoint", portaineree.AuthenticationInternal, testType)

	for i := 0; i < 100; i++ {
		err = store.StoreAuthLog(authLog)
		if err != nil {
			b.Fatalf("Failed adding activity log: %s", err)
		}
	}

	count, err := store.db.Count(&portaineree.AuthActivityLog{})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Number of logs: %d\n", count)

}

func TestAddActivity(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()
	authLog := createAuthLog("username", "endpoint", portaineree.AuthenticationInternal, testType)

	err = store.StoreAuthLog(authLog)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	count, err := store.db.Count(&portaineree.AuthActivityLog{})
	if err != nil {
		t.Fatalf("Failed counting activities: %s", err)
	}

	assert.Equal(t, 1, count, "Store should have one element")
}

func TestGetAuthLogs(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	logs, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{})
	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.AuthActivityLog{log1, log2, log3}, logs)
}

func TestGetAuthLogsByTimestamp(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	time.Sleep(time.Second * 1)

	log2 := createAuthLog("username2", "endpoint2", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	time.Sleep(time.Second * 1)

	log3 := createAuthLog("username3", "endpoint3", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	logs, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		UserActivityLogBaseQuery: portaineree.UserActivityLogBaseQuery{
			BeforeTimestamp: log3.Timestamp - 1,
			AfterTimestamp:  log1.Timestamp + 1,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, log2, logs[0], "logs are not equal")
}

func TestGetAuthLogsByKeyword(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	// like
	shouldHaveAllLogs, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		UserActivityLogBaseQuery: portaineree.UserActivityLogBaseQuery{
			Keyword: "username",
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.AuthActivityLog{log1, log2, log3}, shouldHaveAllLogs)

	// username
	shouldHaveOnlyLog1, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		UserActivityLogBaseQuery: portaineree.UserActivityLogBaseQuery{
			Keyword: "username1",
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, log1, shouldHaveOnlyLog1[0])

	// origin
	shouldHaveOnlyLog3, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		UserActivityLogBaseQuery: portaineree.UserActivityLogBaseQuery{
			Keyword: "endpoint3",
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, log3, shouldHaveOnlyLog3[0])
}

func TestGetAuthLogsByContext(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portaineree.AuthenticationLDAP, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portaineree.AuthenticationOAuth, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	// one type
	shouldHaveLog2, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		ContextTypes: []portaineree.AuthenticationMethod{
			portaineree.AuthenticationLDAP,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.AuthActivityLog{log2}, shouldHaveLog2)

	// two types
	shouldHaveLog1And3, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		ContextTypes: []portaineree.AuthenticationMethod{
			portaineree.AuthenticationInternal,
			portaineree.AuthenticationOAuth,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.AuthActivityLog{log1, log3}, shouldHaveLog1And3)
}

func TestGetAuthLogsByType(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portaineree.AuthenticationInternal, portaineree.AuthenticationActivityFailure)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portaineree.AuthenticationLDAP, portaineree.AuthenticationActivityLogOut)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portaineree.AuthenticationOAuth, portaineree.AuthenticationActivitySuccess)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	// one type
	shouldHaveLog2, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		ActivityTypes: []portaineree.AuthenticationActivityType{
			portaineree.AuthenticationActivityLogOut,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.AuthActivityLog{log2}, shouldHaveLog2)

	// two types
	shouldHaveLog1And3, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		ActivityTypes: []portaineree.AuthenticationActivityType{
			portaineree.AuthenticationActivityFailure,
			portaineree.AuthenticationActivitySuccess,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.AuthActivityLog{log1, log3}, shouldHaveLog1And3)
}

func TestGetAuthLogsSortOrderAndPaginate(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log4 := createAuthLog("username4", "endpoint4", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log4)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	shouldBeLog4AndLog3, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		UserActivityLogBaseQuery: portaineree.UserActivityLogBaseQuery{
			SortDesc: true,
			SortBy:   "Username",
			Offset:   0,
			Limit:    2,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.AuthActivityLog{log4, log3}, shouldBeLog4AndLog3)

	shouldBeLog2AndLog1, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		UserActivityLogBaseQuery: portaineree.UserActivityLogBaseQuery{
			SortDesc: true,
			SortBy:   "Username",
			Offset:   2,
			Limit:    2,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.AuthActivityLog{log2, log1}, shouldBeLog2AndLog1)
}

func TestGetAuthLogsDesc(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portaineree.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	logs, _, err := store.GetAuthLogs(portaineree.AuthLogsQuery{
		UserActivityLogBaseQuery: portaineree.UserActivityLogBaseQuery{
			SortDesc: true,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portaineree.AuthActivityLog{log3, log2, log1}, logs)
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}
