package besticon

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/golang/groupcache"
)

var iconCache *groupcache.Group

type result struct {
	Icons []Icon
	Error string
}

func resultFromCache(siteURL string) ([]Icon, error) {
	if iconCache == nil {
		return FetchIcons(siteURL, false)
	}

	var data []byte
	err := iconCache.Get(nil, siteURL, groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		logger.Println("ERR:", err)
		return FetchIcons(siteURL, false)
	}

	res := &result{}
	err = json.Unmarshal(data, res)
	if err != nil {
		panic(err)
	}

	if res.Error != "" {
		return res.Icons, errors.New(res.Error)
	}
	return res.Icons, nil
}

func generatorFunc(ctx groupcache.Context, siteURL string, sink groupcache.Sink) error {
	icons, err := FetchIcons(siteURL, false)

	res := result{Icons: icons}
	if err != nil {
		res.Error = err.Error()
	}
	bytes, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	sink.SetBytes(bytes)

	return nil
}

func cacheEnabled() bool {
	return iconCache != nil
}

// SetCacheMaxSize enables icon caching if sizeInMB > 0.
func SetCacheMaxSize(sizeInMB int64) {
	if sizeInMB > 0 {
		iconCache = groupcache.NewGroup("icons", sizeInMB<<20, groupcache.GetterFunc(generatorFunc))
	} else {
		iconCache = nil
	}
}

// GetCacheStats returns cache statistics.
func GetCacheStats() groupcache.CacheStats {
	return iconCache.CacheStats(groupcache.MainCache)
}
