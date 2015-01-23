package g_test

import (
	"fmt"
	"github.com/name5566/leaf/go"
)

func Example() {
	d := g.New(10)

	// go 1
	d.Go(func() {
		fmt.Print("Hello ")
	}, func() {
		fmt.Println("World")
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
	// Hello World
	// My name is Leaf
}
