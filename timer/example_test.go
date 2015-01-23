package timer_test

import (
	"fmt"
	"github.com/name5566/leaf/timer"
)

func Example() {
	d := timer.NewDispatcher(10)

	// timer 1
	d.AfterFunc(1, func() {
		fmt.Println("My name is Leaf")
	})

	// timer 2
	t := d.AfterFunc(1, func() {
		fmt.Println("will not print")
	})
	t.Stop()

	// dispatch
	(<-d.ChanTimer).Cb()

	// Output:
	// My name is Leaf
}
