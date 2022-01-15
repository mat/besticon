package besticon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang/groupcache"
)

const contextKeySiteURL = "siteURL"

type result struct {
	Icons []Icon
	Error string
}

func (b *Besticon) resultFromCache(siteURL string) ([]Icon, error) {
	if b.iconCache == nil {
		return b.fetchIcons(siteURL)
	}

	c := context.WithValue(context.Background(), contextKeySiteURL, siteURL)
	var data []byte
	err := b.iconCache.Get(c, cacheKey(siteURL), groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		b.logger.LogError(fmt.Errorf("failed to get icon from cache: %w", err))
		return b.fetchIcons(siteURL)
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

func cacheKey(siteURL string) string {
	// Let results expire after a day
	now := time.Now()
	return fmt.Sprintf("%d-%02d-%02d-%s", now.Year(), now.Month(), now.Day(), siteURL)
}

func (b *Besticon) generatorFunc(ctx context.Context, key string, sink groupcache.Sink) error {
	siteURL := ctx.Value(contextKeySiteURL).(string)
	icons, err := b.fetchIcons(siteURL)
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

type cacheOption struct {
	size int64
}

func (c *cacheOption) applyOption(b *Besticon) {
	b.iconCache = groupcache.NewGroup("icons", c.size<<20, groupcache.GetterFunc(b.generatorFunc))
}

func WithCache(sizeInMB int64) Option {
	return &cacheOption{
		size: sizeInMB,
	}
}

func (b *Besticon) CacheEnabled() bool {
	return b.iconCache != nil
}

// GetCacheStats returns cache statistics.
func (b *Besticon) GetCacheStats() groupcache.CacheStats {
	return b.iconCache.CacheStats(groupcache.MainCache)
}
