package util_test

import (
	"fmt"
	"github.com/name5566/leaf/util"
	"sync"
)

func ExampleMap() {
	m := new(util.Map)

	fmt.Println(m.Get("key"))
	m.Set("key", "value")
	fmt.Println(m.Get("key"))
	m.Del("key")
	fmt.Println(m.Get("key"))

	m.Set(1, "1")
	m.Set(2, 2)
	m.Set("3", 3)

	fmt.Println(m.Len())

	// Output:
	// <nil>
	// value
	// <nil>
	// 3
}
