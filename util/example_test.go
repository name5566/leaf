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

func ExampleCallRouter() {
	r := util.NewCallRouter(10)
	var wg sync.WaitGroup

	// goroutine 1
	wg.Add(1)
	go func() {
		// def
		r.Def("add", func(args []interface{}) interface{} {
			num1 := args[0].(int)
			num2 := args[1].(int)
			return num1 + num2
		})

		// route
		ci := <-r.Chan()
		r.Route(ci)

		wg.Done()
	}()

	// goroutine 2
	wg.Add(1)
	go func() {
		// call
		c := r.Call1("add", 1, 2)
		fmt.Println(<-c)

		wg.Done()
	}()

	wg.Wait()

	// Output:
	// 3
}
