package timer_test

import (
	"fmt"
	"github.com/name5566/leaf/timer"
)

func Example() {
	d := timer.NewDispatcher(10)

	var counter int
	for i := 0; i < 10000; i++ {
		d.AfterFunc(0, func() {
			counter++
		})

		(<-d.ChanCb)()
	}

	fmt.Println(counter)

	// Output:
	// 10000
}
