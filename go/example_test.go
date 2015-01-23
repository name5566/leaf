package g_test

import (
	"fmt"
	"github.com/name5566/leaf/go"
)

func Example() {
	d := g.New(10)

	d.Go(func() {
		fmt.Print("My name is ")
	}, func() {
		fmt.Println("Leaf")
	})

	d.Close()

	// Output:
	// My name is Leaf
}
