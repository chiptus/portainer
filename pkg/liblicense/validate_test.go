package liblicense

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestExpired(t *testing.T) {
	tests := map[string]struct {
		created      int64
		expiredAfter int
		want         bool
	}{
		"expired": {created: time.Now().Unix() - 3600*24*32, expiredAfter: 30, want: true},
		"lastDay": {created: time.Now().Unix() - 3600*24*30, expiredAfter: 30, want: false},
		"valid":   {created: time.Now().Unix() - 3600*24*20, expiredAfter: 30, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := Expired(tc.created, tc.expiredAfter)
			if tc.want != got {
				t.Fatalf("expected: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestEndOfDay(t *testing.T) {
	t1 := time.Now()
	t2 := EndOfDay(t1)
	expected := fmt.Sprintf("%s %02d 23:59:59", t1.Month(), t1.Day())
	got := t2.Format("January 02 15:04:05")
	if got != expected {
		t.Fatalf("expected: %v, got: %v", expected, got)
	}
}

func TestTrimError(t *testing.T) {
	err1 := fmt.Errorf("Invalid response status: %d, body: %s", 502, "body")
	err2 := fmt.Errorf("Invalid response status: %d, body: %s", 400, "body")
	err3 := fmt.Errorf("Invalid response status: %d, body: %s", 500, "body")
	err4 := errors.New("Post " + "\"test.api.com\"" + ": dial tcp: lookup test.api.com: no such host")
	table := map[error]error{
		err1: nil,
		err2: err2,
		err3: err3,
		err4: nil,
	}
	for k, v := range table {
		if result := trimError(k); result != v {
			t.Errorf("testing: %v, want: %v, got: %v", k, v, result)
		}
	}
}
