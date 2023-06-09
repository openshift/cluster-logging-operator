package utils

import "sync"

// StringMap is a concurrent-safe map[string]string based on sync.Map
type StringMap struct {
	m sync.Map
}
type Float64Map struct {
	m sync.Map
}

func InitStringMap(m map[string]string) *StringMap {
	sm := &StringMap{}
	for key, val := range m {
		sm.Set(key, val)
	}
	return sm
}

func InitFloat64Map(m map[string]float64) *Float64Map {
	fm := &Float64Map{}
	for key, val := range m {
		fm.m.Store(key, val)
	}
	return fm
}

// Get is a wrapper for getting the value from the underlying map
func (sm *StringMap) Get(key string) string {
	if val, ok := sm.m.Load(key); ok {
		return val.(string)
	}
	return ""
}

// Set is a wrapper for setting the value of a key in the underlying map
func (sm *StringMap) Set(key string, val string) {
	sm.m.Store(key, val)
}

// for prometheus gauge metric setting up a value of float64
func (fm *Float64Map) Inc(key string) {
	if val, ok := fm.m.Load(key); ok {
		f := val.(float64) + 1
		fm.m.Store(key, f)
	}
}

// for prometheus gauge metric getting a value of float64
func (fm *Float64Map) Get(key string) float64 {
	if val, ok := fm.m.Load(key); ok {
		return val.(float64)
	}
	return 0
}

func (sm *StringMap) Range(f func(key, value interface{}) bool) {
	sm.m.Range(f)
}

func (fm *Float64Map) Range(f func(key, value interface{}) bool) {
	fm.m.Range(f)
}
