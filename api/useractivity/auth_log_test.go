package useractivity

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	portainer "github.com/portainer/portainer/api"
)

const (
	testType = portainer.AuthenticationActivityType(0)
)

func createAuthLog(username, origin string, context portainer.AuthenticationMethod, activityType portainer.AuthenticationActivityType) *portainer.AuthActivityLog {
	return &portainer.AuthActivityLog{
		Type: activityType,
		UserActivityLogBase: portainer.UserActivityLogBase{
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

	authLog := createAuthLog("username", "endpoint", portainer.AuthenticationInternal, testType)

	for i := 0; i < 100; i++ {
		err = store.StoreAuthLog(authLog)
		if err != nil {
			b.Fatalf("Failed adding activity log: %s", err)
		}
	}

	count, err := store.db.Count(&portainer.AuthActivityLog{})
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
	authLog := createAuthLog("username", "endpoint", portainer.AuthenticationInternal, testType)

	err = store.StoreAuthLog(authLog)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	count, err := store.db.Count(&portainer.AuthActivityLog{})
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

	log1 := createAuthLog("username1", "endpoint1", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	logs, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{})
	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portainer.AuthActivityLog{log1, log2, log3}, logs)
}

func TestGetAuthLogsByTimestamp(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	time.Sleep(time.Second * 1)

	log2 := createAuthLog("username2", "endpoint2", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	time.Sleep(time.Second * 1)

	log3 := createAuthLog("username3", "endpoint3", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	logs, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		UserActivityLogBaseQuery: portainer.UserActivityLogBaseQuery{
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

	log1 := createAuthLog("username1", "endpoint1", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	// like
	shouldHaveAllLogs, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		UserActivityLogBaseQuery: portainer.UserActivityLogBaseQuery{
			Keyword: "username",
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portainer.AuthActivityLog{log1, log2, log3}, shouldHaveAllLogs)

	// username
	shouldHaveOnlyLog1, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		UserActivityLogBaseQuery: portainer.UserActivityLogBaseQuery{
			Keyword: "username1",
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, log1, shouldHaveOnlyLog1[0])

	// origin
	shouldHaveOnlyLog3, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		UserActivityLogBaseQuery: portainer.UserActivityLogBaseQuery{
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

	log1 := createAuthLog("username1", "endpoint1", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portainer.AuthenticationLDAP, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portainer.AuthenticationOAuth, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	// one type
	shouldHaveLog2, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		ContextTypes: []portainer.AuthenticationMethod{
			portainer.AuthenticationLDAP,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portainer.AuthActivityLog{log2}, shouldHaveLog2)

	// two types
	shouldHaveLog1And3, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		ContextTypes: []portainer.AuthenticationMethod{
			portainer.AuthenticationInternal,
			portainer.AuthenticationOAuth,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portainer.AuthActivityLog{log1, log3}, shouldHaveLog1And3)
}

func TestGetAuthLogsByType(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portainer.AuthenticationInternal, portainer.AuthenticationActivityFailure)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portainer.AuthenticationLDAP, portainer.AuthenticationActivityLogOut)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portainer.AuthenticationOAuth, portainer.AuthenticationActivitySuccess)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	// one type
	shouldHaveLog2, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		ActivityTypes: []portainer.AuthenticationActivityType{
			portainer.AuthenticationActivityLogOut,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portainer.AuthActivityLog{log2}, shouldHaveLog2)

	// two types
	shouldHaveLog1And3, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		ActivityTypes: []portainer.AuthenticationActivityType{
			portainer.AuthenticationActivityFailure,
			portainer.AuthenticationActivitySuccess,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portainer.AuthActivityLog{log1, log3}, shouldHaveLog1And3)
}

func TestGetAuthLogsSortOrderAndPaginate(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log4 := createAuthLog("username4", "endpoint4", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log4)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	shouldBeLog4AndLog3, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		UserActivityLogBaseQuery: portainer.UserActivityLogBaseQuery{
			SortDesc: true,
			SortBy:   "Username",
			Offset:   0,
			Limit:    2,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portainer.AuthActivityLog{log4, log3}, shouldBeLog4AndLog3)

	shouldBeLog2AndLog1, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		UserActivityLogBaseQuery: portainer.UserActivityLogBaseQuery{
			SortDesc: true,
			SortBy:   "Username",
			Offset:   2,
			Limit:    2,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portainer.AuthActivityLog{log2, log1}, shouldBeLog2AndLog1)
}

func TestGetAuthLogsDesc(t *testing.T) {
	store, err := setup(t.TempDir())
	if err != nil {
		t.Fatalf("Failed setup: %s", err)
	}

	defer store.Close()

	log1 := createAuthLog("username1", "endpoint1", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log1)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log2 := createAuthLog("username2", "endpoint2", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log2)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	log3 := createAuthLog("username3", "endpoint3", portainer.AuthenticationInternal, testType)
	err = store.StoreAuthLog(log3)
	if err != nil {
		t.Fatalf("Failed adding activity log: %s", err)
	}

	logs, _, err := store.GetAuthLogs(portainer.AuthLogsQuery{
		UserActivityLogBaseQuery: portainer.UserActivityLogBaseQuery{
			SortDesc: true,
		},
	})

	if err != nil {
		t.Fatalf("failed fetching logs: %s", err)
	}

	assert.Equal(t, []*portainer.AuthActivityLog{log3, log2, log1}, logs)
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}
