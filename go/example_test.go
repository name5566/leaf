package g_test

import (
	"fmt"
	"github.com/name5566/leaf/go"
	"time"
)

func Example() {
	d := g.New(10)

	// go 1
	var res int
	d.Go(func() {
		fmt.Println("1 + 1 = ?")
		res = 1 + 1
	}, func() {
		fmt.Println(res)
	})

	d.Cb(<-d.ChanCb)

	// go 2
	d.Go(func() {
		fmt.Print("My name is ")
	}, func() {
		fmt.Println("Leaf")
	})

	d.Close()

	// Output:
	// 1 + 1 = ?
	// 2
	// My name is Leaf
}

func ExampleLinearContext() {
	d := g.New(10)

	// parallel
	d.Go(func() {
		time.Sleep(time.Second / 2)
		fmt.Println("1")
	}, nil)
	d.Go(func() {
		fmt.Println("2")
	}, nil)

	d.Cb(<-d.ChanCb)
	d.Cb(<-d.ChanCb)

	// linear
	c := d.NewLinearContext()
	c.Go(func() {
		time.Sleep(time.Second / 2)
		fmt.Println("1")
	}, nil)
	c.Go(func() {
		fmt.Println("2")
	}, nil)

	d.Close()

	// Output:
	// 2
	// 1
	// 1
	// 2
}
