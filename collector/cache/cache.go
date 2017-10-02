package cache

import "fmt"

func NewCache() Cache {
	return &Caches{
		data: make(map[interface{}]record),
	}
}

func (c *Caches) Add(u interface{}, value interface{}) (err error) {

	c.Lock()
	defer c.Unlock()

	if _, ok := c.data[u]; !ok {

		c.data[u] = record{
			value: value,
		}
		return nil
	}

	return fmt.Errorf("Item Exists - Use update")
}

// Get retrieves the entry from the cache
func (c *Caches) Get(u interface{}) (i interface{}, err error) {

	c.Lock()
	defer c.Unlock()

	if _, ok := c.data[u]; !ok {
		return "", fmt.Errorf("Item does not exist")
	}

	return c.data[u], nil
}
