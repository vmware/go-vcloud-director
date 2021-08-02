package util

import (
	"log"
	"sync"
)

// Imported from Hashicorp (https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html)

// MutexKV is a simple key/value store for arbitrary mutexes. It can be used to
// serialize changes across arbitrary collaborators that share knowledge of the
// keys they must serialize on.
//
// The initial use case is to let aws_security_group_rule resources serialize
// their access to individual security groups based on SG ID.
type MutexKV struct {
	lock   sync.Mutex
	store  map[string]*sync.Mutex
	silent bool
}

// Lock locks the mutex for the given key. Caller is responsible for calling Unlock
// for the same key
func (m *MutexKV) Lock(key string) {
	if !m.silent {
		log.Printf("[DEBUG] Locking %q", key)
	}
	m.get(key).Lock()
	if !m.silent {
		log.Printf("[DEBUG] Locked %q", key)
	}
}

// Unlock unlocks the mutex for the given key. Caller must have called Lock for the same key first
func (m *MutexKV) Unlock(key string) {
	if !m.silent {
		log.Printf("[DEBUG] Unlocking %q", key)
	}
	m.get(key).Unlock()
	if !m.silent {
		log.Printf("[DEBUG] Unlocked %q", key)
	}
}

// get returns a mutex for the given key, no guarantee of its lock status
func (m *MutexKV) get(key string) *sync.Mutex {
	m.lock.Lock()
	defer m.lock.Unlock()
	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}
	return mutex
}

// NewMutexKV returns a properly initalized MutexKV
func NewMutexKV() *MutexKV {
	return &MutexKV{
		store: make(map[string]*sync.Mutex),
	}
}

// NewMutexKVSilent returns a properly initalized MutexKV with the silent property set
func NewMutexKVSilent() *MutexKV {
	return &MutexKV{
		store:  make(map[string]*sync.Mutex),
		silent: true,
	}
}
