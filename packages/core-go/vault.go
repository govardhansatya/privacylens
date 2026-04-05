package privacylens

import "fmt"

type SessionVault interface {
	Store(sessionID, token, value string)
	Retrieve(sessionID, token string) (string, error)
	Clear(sessionID string)
}

type MemoryVault struct {
	data map[string]map[string]string
}

func NewMemoryVault() *MemoryVault {
	return &MemoryVault{data: map[string]map[string]string{}}
}

func (v *MemoryVault) Store(sessionID, token, value string) {
	session, ok := v.data[sessionID]
	if !ok {
		session = map[string]string{}
		v.data[sessionID] = session
	}
	session[token] = value
}

func (v *MemoryVault) Retrieve(sessionID, token string) (string, error) {
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
	delete(v.data, sessionID)
}
