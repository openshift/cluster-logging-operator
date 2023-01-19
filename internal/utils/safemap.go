package utils

import "sync"

// StringMap is a concurrent-safe map[string]string based on sync.Map
type StringMap struct {
	sync.RWMutex
	M map[string]string
}
type Float64Map struct {
	sync.RWMutex
	M map[string]float64
}

// Get is a wrapper for getting the value from the underlying map
func (r *StringMap) Get(key string) string {
	r.RLock()
	defer r.RUnlock()
	return r.M[key]
}

// Set is a wrapper for setting the value of a key in the underlying map
func (r *StringMap) Set(key string, val string) {
	r.Lock()
	defer r.Unlock()
	r.M[key] = val
}

// for prometheus gauge metric setting up a value of float64
func (r *Float64Map) Inc(key string) {
	r.Lock()
	defer r.Unlock()
	r.M[key]++
}

// for prometheus gauge metric getting a value of float64
func (f *Float64Map) Get(key string) float64 {
	f.RLock()
	defer f.RUnlock()
	return f.M[key]
}
