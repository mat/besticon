package besticon

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang/groupcache"
)

var iconCache *groupcache.Group

type result struct {
	Icons []Icon
	Error string
}

func resultFromCache(siteURL string) ([]Icon, error) {
	if iconCache == nil {
		return fetchIcons(siteURL)
	}

	// Let results expire after a day
	now := time.Now()
	key := fmt.Sprintf("%d-%02d-%02d-%s", now.Year(), now.Month(), now.Day(), siteURL)
	var data []byte
	err := iconCache.Get(siteURL, key, groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		logger.Println("ERR:", err)
		return fetchIcons(siteURL)
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

func generatorFunc(ctx groupcache.Context, key string, sink groupcache.Sink) error {
	siteURL := ctx.(string)
	icons, err := fetchIcons(siteURL)
	if err != nil {
		// Don't cache errors
		return err
	}

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

func CacheEnabled() bool {
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
