package webdav

import (
	"context"
	"overlink.top/app/system/conf"
	"overlink.top/app/system/msg"
	"strings"
	"sync"
	"time"
)

// FileInfoCache caches file information to eliminate double lookups
type FileInfoCache struct {
	cache   map[string]*cachedItem
	mutex   sync.RWMutex
	ttl     time.Duration
	maxSize int
	order   []string // Track insertion order for LRU eviction
}

// cachedItem wraps cached data with expiration time
type cachedItem struct {
	fileInfo   msg.Finfo        // For individual files
	dirList    []msg.Finfo      // For directory listings
	isDirList  bool             // Flag to indicate if this is a directory listing
	expireTime time.Time
}

// NewFileInfoCache creates a new file info cache with the specified TTL and max size
func NewFileInfoCache(ttl time.Duration, maxSize int) *FileInfoCache {
	return &FileInfoCache{
		cache:   make(map[string]*cachedItem),
		ttl:     ttl,
		maxSize: maxSize,
		order:   make([]string, 0),
	}
}

// GetFile retrieves file info from cache if it exists and hasn't expired
func (c *FileInfoCache) GetFile(path string) (msg.Finfo, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if item, exists := c.cache[path]; exists && !item.isDirList {
		if time.Now().Before(item.expireTime) {
			return item.fileInfo, true
		}
		// Item expired, remove it
		go c.remove(path)
	}

	return nil, false
}

// GetDirList retrieves directory listing from cache if it exists and hasn't expired
func (c *FileInfoCache) GetDirList(path string) ([]msg.Finfo, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if item, exists := c.cache[path]; exists && item.isDirList {
		if time.Now().Before(item.expireTime) {
			return item.dirList, true
		}
		// Item expired, remove it
		go c.remove(path)
	}

	return nil, false
}

// SetFile stores file info in cache with expiration
func (c *FileInfoCache) SetFile(path string, info msg.Finfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Remove the path from order if it already exists
	c.removeFromOrder(path)

	// Add to cache
	c.cache[path] = &cachedItem{
		fileInfo:   info,
		isDirList:  false,
		expireTime: time.Now().Add(c.ttl),
	}

	// Add to order
	c.order = append(c.order, path)

	// Evict oldest items if cache is too large
	c.evictIfNeeded()
}

// SetDirList stores directory listing in cache with expiration
func (c *FileInfoCache) SetDirList(path string, list []msg.Finfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Remove the path from order if it already exists
	c.removeFromOrder(path)

	// Add to cache
	c.cache[path] = &cachedItem{
		dirList:    list,
		isDirList:  true,
		expireTime: time.Now().Add(c.ttl),
	}

	// Add to order
	c.order = append(c.order, path)

	// Evict oldest items if cache is too large
	c.evictIfNeeded()
}

// Remove removes an item from cache
func (c *FileInfoCache) remove(path string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, path)
}

// Clear removes all items from cache
func (c *FileInfoCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*cachedItem)
	c.order = make([]string, 0)
}

// removeFromOrder removes a path from the order slice
func (c *FileInfoCache) removeFromOrder(path string) {
	for i, p := range c.order {
		if p == path {
			// Remove the element at index i
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

// evictIfNeeded removes the oldest items if cache exceeds maxSize
func (c *FileInfoCache) evictIfNeeded() {
	if c.maxSize <= 0 {
		return // No size limit
	}

	for len(c.cache) > c.maxSize && len(c.order) > 0 {
		// Remove the oldest item (first in order)
		oldestPath := c.order[0]
		delete(c.cache, oldestPath)
		c.order = c.order[1:]
	}
}

// Invalidate removes a specific item from cache
func (c *FileInfoCache) Invalidate(path string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, path)
	c.removeFromOrder(path)
}

// InvalidatePattern removes items from cache that match a pattern
func (c *FileInfoCache) InvalidatePattern(pattern string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	pathsToRemove := make([]string, 0)
	for path := range c.cache {
		if strings.Contains(path, pattern) {
			pathsToRemove = append(pathsToRemove, path)
		}
	}

	for _, path := range pathsToRemove {
		delete(c.cache, path)
		c.removeFromOrder(path)
	}
}

// Global cache instance with configurable TTL
var globalFileInfoCache *FileInfoCache

func init() {
	// Initialize cache with configured TTL or default 5 minutes
	ttl := 5 * time.Minute
	if conf.AppConf.WebDAV.MetadataCacheTTL > 0 {
		ttl = time.Duration(conf.AppConf.WebDAV.MetadataCacheTTL) * time.Second
	}

	// Initialize cache with configured size or default 10000 items
	maxSize := 10000
	if conf.AppConf.WebDAV.CacheSize > 0 {
		maxSize = conf.AppConf.WebDAV.CacheSize
	}

	globalFileInfoCache = NewFileInfoCache(ttl, maxSize)
}

// GetCachedFileInfo retrieves file info from cache or fetches it
func GetCachedFileInfo(ctx context.Context, path string) (msg.Finfo, error) {
	// Try to get from cache first
	if info, found := globalFileInfoCache.GetFile(path); found {
		return info, nil
	}

	// Not in cache, fetch it (this would be implemented based on your logic)
	// For now, we'll return nil to indicate it wasn't found in cache
	return nil, nil
}

// CacheFileInfo stores file info in cache
func CacheFileInfo(path string, info msg.Finfo) {
	globalFileInfoCache.SetFile(path, info)
}

// GetCachedDirList retrieves directory listing from cache or fetches it
func GetCachedDirList(ctx context.Context, path string) ([]msg.Finfo, error) {
	// Try to get from cache first
	if list, found := globalFileInfoCache.GetDirList(path); found {
		return list, nil
	}

	// Not in cache, fetch it (this would be implemented based on your logic)
	// For now, we'll return nil to indicate it wasn't found in cache
	return nil, nil
}

// CacheDirList stores directory listing in cache
func CacheDirList(path string, list []msg.Finfo) {
	globalFileInfoCache.SetDirList(path, list)
}

// InvalidateCache removes a specific item from cache
func InvalidateCache(path string) {
	globalFileInfoCache.Invalidate(path)
}

// InvalidateCachePattern removes items from cache that match a pattern
func InvalidateCachePattern(pattern string) {
	globalFileInfoCache.InvalidatePattern(pattern)
}