package plist

import (
	"os"
	"sync"
	"time"
)

// Cache provides mtime-based caching for parsed plist files.
// It avoids re-parsing unchanged files on every refresh cycle.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
}

type cacheEntry struct {
	data    PlistData
	modTime time.Time
}

// NewCache creates an empty plist cache.
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]*cacheEntry),
	}
}

// Get returns the cached PlistData for path, re-parsing only if the file's
// mtime has changed since the last read.
func (c *Cache) Get(path string) (*PlistData, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	entry, ok := c.entries[path]
	c.mu.RUnlock()

	if ok && entry.modTime.Equal(info.ModTime()) {
		return &entry.data, nil
	}

	// Cache miss — re-parse
	data, err := ReadPlist(path)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.entries[path] = &cacheEntry{data: *data, modTime: info.ModTime()}
	c.mu.Unlock()

	return data, nil
}
