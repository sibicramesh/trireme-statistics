package cache

import "fmt"

func NewCache() Cache {
	return &Caches{
		data: make(map[interface{}]string),
	}
}

func (c *Caches) Add(u interface{}, value string) (err error) {

	c.Lock()
	defer c.Unlock()

	if _, ok := c.data[u]; !ok {

		c.data[u] = value
		return nil
	}

	return fmt.Errorf("Item Exists - Use update")
}

// Get retrieves the entry from the cache
func (c *Caches) Get(u interface{}) (i string, err error) {

	c.Lock()
	defer c.Unlock()

	if _, ok := c.data[u]; !ok {
		return "", fmt.Errorf("Item does not exist")
	}

	return c.data[u], nil
}
