// Package storage takes care of storing the actual values in memory
package storage

import (
	"sync"
)

// GenericConcurrentMap maps a string key to a int or string value
type GenericConcurrentMap struct {
	sync.RWMutex
	internal map[string]string
}

// NewGenericConcurrentMap creates a new string > int or string map
func NewGenericConcurrentMap() *GenericConcurrentMap {
	gcm := GenericConcurrentMap{
		internal: make(map[string]string),
	}
	return &gcm
}

// Get new value from the map or nil, if it does not exist
func (gcm *GenericConcurrentMap) Get(key string) (value string, ok bool) {
	gcm.RLock()
	defer gcm.RUnlock()
	result, ok := gcm.internal[key]
	return result, ok
}

// Delete value at a given key, and returns true if deleted, false otherwise
func (gcm *GenericConcurrentMap) Delete(key string) bool {
	gcm.Lock()
	defer gcm.Unlock()
	_, ok := gcm.internal[key]
	if ok == false {
		return false
	}
	delete(gcm.internal, key)
	return true
}

// Set a given int or string value at given key
func (gcm *GenericConcurrentMap) Set(key string, value string) {
	gcm.Lock()
	defer gcm.Unlock()
	gcm.internal[key] = value
}
