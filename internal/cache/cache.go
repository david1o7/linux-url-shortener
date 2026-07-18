package cache

import "sync"

type Cache struct{
	data map[string]string
	mu sync.RWMutex
}

func NewCache() *Cache{
	return &Cache{
		data: make(map[string]string),
	}
}

func (c *Cache) Get(code string) (string, bool){
	c.mu.RLock()

	defer c.mu.RUnlock()

	url, ok := c.data[code]

	return url, ok
}

func (c *Cache) Set(code string, url string){
	
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[code] = url
}