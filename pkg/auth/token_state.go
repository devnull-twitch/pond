package auth

import (
	"sync"
	"time"

	rpc_ponds "github.com/devnull-twitch/pond-com/protobuf/com/v1"
)

type tokenStorageEntry struct {
	status rpc_ponds.PollResponse_Status
	ts     time.Time
	jwt    string
}

var (
	m           sync.Mutex
	reqTokenMap map[string]tokenStorageEntry
)

func Add(token string) bool {
	if _, exists := reqTokenMap[token]; exists {
		return false
	}

	m.Lock()
	reqTokenMap[token] = tokenStorageEntry{
		status: rpc_ponds.PollResponse_STATUS_WAITING,
		ts:     time.Now(),
	}
	m.Unlock()

	go cleanup()
	return true
}

func Set(token string, status rpc_ponds.PollResponse_Status, jwt *string) bool {
	if _, exists := reqTokenMap[token]; !exists {
		return false
	}

	t := ""
	if jwt != nil {
		t = *jwt
	}

	m.Lock()
	reqTokenMap[token] = tokenStorageEntry{
		status: status,
		ts:     time.Now(),
		jwt:    t,
	}
	m.Unlock()

	go cleanup()
	return true
}

func Get(token string) (rpc_ponds.PollResponse_Status, string, bool) {
	if _, exists := reqTokenMap[token]; !exists {
		return rpc_ponds.PollResponse_STATUS_UNSPECIFIED, "", false
	}

	go cleanup()
	return reqTokenMap[token].status, reqTokenMap[token].jwt, true
}

func cleanup() {
	m.Lock()
	for key, data := range reqTokenMap {
		if data.ts.Before(time.Now().Add(-time.Hour)) {
			delete(reqTokenMap, key)
		}
	}
	m.Unlock()
}
