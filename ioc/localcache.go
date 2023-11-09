// Copyright@daidai53 2023
package ioc

import cache "github.com/daidai53/localcache"

func NewLocalCacheDefault() cache.LocalCache {
	return cache.NewLocalCacheV1(64)
}
