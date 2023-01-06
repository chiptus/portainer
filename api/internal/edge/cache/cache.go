package cache

import (
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/VictoriaMetrics/fastcache"
)

var c = fastcache.New(1)

func key(k portaineree.EndpointID) []byte {
	return []byte(strconv.Itoa(int(k)))
}

func Set(k portaineree.EndpointID, v []byte) {
	c.Set(key(k), v)
}

func Get(k portaineree.EndpointID) ([]byte, bool) {
	return c.HasGet(nil, key(k))
}

func Del(k portaineree.EndpointID) {
	c.Del(key(k))
}
