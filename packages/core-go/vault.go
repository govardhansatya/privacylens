package privacylens

import (
	"fmt"
	"sync"
)

type SessionVault interface {
	Store(sessionID, token, value string)
	Retrieve(sessionID, token string) (string, error)
	Clear(sessionID string)
}

type MemoryVault struct {
	mu   sync.RWMutex
	data map[string]map[string]string
}

func NewMemoryVault() *MemoryVault {
	return &MemoryVault{data: map[string]map[string]string{}}
}

func (v *MemoryVault) Store(sessionID, token, value string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	session, ok := v.data[sessionID]
	if !ok {
		session = map[string]string{}
		v.data[sessionID] = session
	}
	session[token] = value
}

func (v *MemoryVault) Retrieve(sessionID, token string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	session, ok := v.data[sessionID]
	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}
	value, ok := session[token]
	if !ok {
		return "", fmt.Errorf("token not found in session: %s", token)
	}
	return value, nil
}

func (v *MemoryVault) Clear(sessionID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.data, sessionID)
}
