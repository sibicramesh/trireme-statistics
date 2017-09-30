package cache

import "fmt"

func init() {
	nc := NewCache()
	nc.Add("ContextID", "abcdefg")
	fmt.Println(nc.Get("ContextID"))
}
