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

func ExampleRandIntervalN() {
	r := util.RandIntervalN(-1, 0, 2)
	if r[0] == -1 && r[1] == 0 ||
		r[0] == 0 && r[1] == -1 {
		fmt.Println("ok")
	}

	// Output:
	// ok
}

func ExampleDeepCopy() {
	src := []int{1, 2, 3}

	var dst []int
	util.DeepCopy(&dst, &src)

	for _, v := range dst {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 2
	// 3
}

func ExampleDeepClone() {
	src := []int{1, 2, 3}

	dst := util.DeepClone(src).([]int)

	for _, v := range dst {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 2
	// 3
}
