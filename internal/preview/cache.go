package preview

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type CacheKey struct {
	Fingerprint string
	Seed        int
	LightMode   bool
	Mode        string
	GroupHash   string
}

func (k CacheKey) String() string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%d|%v|%s|%s", k.Fingerprint, k.Seed, k.LightMode, k.Mode, k.GroupHash)
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

type Cache struct {
	mu   sync.RWMutex
	dir  string
	keys map[string]bool
}

func NewCache(dir string) (*Cache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create cache dir: %w", err)
	}
	return &Cache{
		dir:  dir,
		keys: make(map[string]bool),
	}, nil
}

func (c *Cache) Get(key CacheKey) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	k := key.String()
	if c.keys[k] {
		p := filepath.Join(c.dir, k+".png")
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}

func (c *Cache) Put(key CacheKey, srcPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	k := key.String()
	dst := filepath.Join(c.dir, k+".png")

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return err
	}
	c.keys[k] = true
	return nil
}

func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	os.RemoveAll(c.dir)
	os.MkdirAll(c.dir, 0755)
	c.keys = make(map[string]bool)
}

func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	os.RemoveAll(c.dir)
	c.keys = make(map[string]bool)
}
