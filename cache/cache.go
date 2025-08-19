package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"yuki_zpm.org/logger"
)

type Cache struct {
	cacheDir string
	entries  map[string]Entry
}

type Entry struct {
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
	Version  string `json:"version"`
	CommitSHA string 
}

func New() *Cache {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	
	cacheDir := filepath.Join(homeDir, ".yuki", "cache")
	
	cache := &Cache{
		cacheDir: cacheDir,
		entries:  make(map[string]Entry),
	}
	
	cache.load()
	return cache
}

func (c *Cache) GetCacheDir() string {
	return c.cacheDir
}

func (c *Cache) Get(key string) (Entry, bool) {
	entry, exists := c.entries[key]
	if !exists {
		return Entry{}, false
	}
	
	if _, err := os.Stat(entry.Path); os.IsNotExist(err) {
		delete(c.entries, key)
		c.save()
		return Entry{}, false
	}
	
	return entry, true
}

func (c *Cache) Set(key string, entry Entry) {
	c.entries[key] = entry
	c.save()
}

func (c *Cache) Delete(key string) {
	delete(c.entries, key)
	c.save()
}

func (c *Cache) Clear() error {
	c.entries = make(map[string]Entry)
	
	if err := os.RemoveAll(c.cacheDir); err != nil {
		return err
	}
	
	return c.ensureCacheDir()
}

func (c *Cache) load() {
	if err := c.ensureCacheDir(); err != nil {
		logger.Debug("Failed to create cache directory: %v", err)
		return
	}
	
	indexPath := filepath.Join(c.cacheDir, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Debug("Failed to read cache index: %v", err)
		}
		return
	}
	
	if err := json.Unmarshal(data, &c.entries); err != nil {
		logger.Debug("Failed to parse cache index: %v", err)
	}
}

func (c *Cache) save() {
	if err := c.ensureCacheDir(); err != nil {
		logger.Debug("Failed to create cache directory: %v", err)
		return
	}
	
	indexPath := filepath.Join(c.cacheDir, "index.json")
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		logger.Debug("Failed to marshal cache index: %v", err)
		return
	}
	
	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		logger.Debug("Failed to write cache index: %v", err)
	}
}

func (c *Cache) ensureCacheDir() error {
	return os.MkdirAll(c.cacheDir, 0755)
}

func (c *Cache) GetSize() (int64, error) {
	var size int64
	
	err := filepath.Walk(c.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		size += info.Size()
		return nil
	})
	
	return size, err
}

func (c *Cache) ListEntries() map[string]Entry {
	result := make(map[string]Entry)
	for k, v := range c.entries {
		result[k] = v
	}
	return result
}
