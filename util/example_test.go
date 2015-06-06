package util_test

import (
	"fmt"
	"github.com/name5566/leaf/util"
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

func ExampleRandGroup() {
	i := util.RandGroup(0, 0, 50, 50)
	switch i {
	case 2, 3:
		fmt.Println("ok")
	}

	// Output:
	// ok
}

func ExampleRandInterval() {
	v := util.RandInterval(-1, 1)
	switch v {
	case -1, 0, 1:
		fmt.Println("ok")
	}

	// Output:
	// ok
}

func ExampleDeepClone() {
	type Struct struct {
		Point *int
		Map   map[string]int
		Slice []int
	}
	src := &Struct{
		new(int),
		make(map[string]int),
		[]int{},
	}

	// value
	*(src.Point) = 1
	src.Map["leaf"] = 2
	src.Slice = append(src.Slice, 3)

	// deep clone
	// dst := new(Struct)
	// *dst = *src
	dst := util.DeepClone(src).(*Struct)

	// new value
	*(src.Point) = 10
	src.Map["leaf"] = 20
	src.Slice[0] = 30

	fmt.Println(*(dst.Point))
	fmt.Println(dst.Map["leaf"])
	fmt.Println(dst.Slice[0])

	// Output:
	// 1
	// 2
	// 3
}
