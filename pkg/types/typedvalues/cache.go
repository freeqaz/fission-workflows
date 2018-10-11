package typedvalues

import (
	"runtime"

	"github.com/hashicorp/golang-lru"
)

// TODO add inverse lookup - prevent wrapping of common values

var cache *lru.Cache

func InitCache(n int) error {
	var err error
	cache, err = lru.New(n)
	return err
}

func cachePut(tv *TypedValue, val interface{}) {
	if cache == nil {
		return
	}
	if tv == nil {
		return
	}
	cache.Add(tv, val)
	runtime.SetFinalizer(tv, cacheEvict)
}

func cacheEvict(tv *TypedValue) {
	if cache == nil {
		return
	}
	if tv == nil {
		return
	}
	cache.Remove(tv)
}

func cacheLookup(tv *TypedValue) (interface{}, bool) {
	if cache == nil {
		return nil, false
	}
	if tv == nil {
		return nil, false
	}
	return cache.Get(tv)
}
