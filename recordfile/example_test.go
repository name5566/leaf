package recordfile_test

import (
	"fmt"
	"github.com/name5566/leaf/recordfile"
)

func Example() {
	type Record struct {
		// index 0
		IndexInt int "index"
		Number   int32
		// index 1
		IndexStr string "index"
		Str      string
	}

	rf, err := recordfile.New(Record{})
	if err != nil {
		return
	}

	err = rf.Read("test.txt")
	if err != nil {
		return
	}

	for i := 0; i < rf.NumRecord(); i++ {
		r := rf.Record(i).(*Record)
		fmt.Println(r.IndexInt)
	}

	r := rf.Index(2).(*Record)
	fmt.Println(r.Str)

	r = rf.Indexes(0)[2].(*Record)
	fmt.Println(r.Str)

	r = rf.Indexes(1)["three"].(*Record)
	fmt.Println(r.Str)

	// Output:
	// 1
	// 2
	// 3
	// cat
	// cat
	// book
}
