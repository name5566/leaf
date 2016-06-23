package recordfile

import (
	"testing"
)

	type Mysql struct {
		Host, Port, User, Passwd, Db string
	}

func Test_map(t *testing.T){
	type Record struct {
		// index 0
		IndexInt int "index"
		// index 1
		IndexStr string "index"
		_Number  int32
		Str      string
		Arr1     [2]int
		Arr2     [3][2]int
		Arr3     []int
		St       struct {
			Name string "name"
			Num  int    "num"
		}
		Maps map[string]Mysql
	}

	rf, err := New(Record{})
	if err != nil {
		t.Error(err)
	}

	err = rf.Read("test.txt")
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < rf.NumRecord(); i++ {
		r := rf.Record(i).(*Record)
		t.Log(r.IndexInt)
	}

	r := rf.Index(2).(*Record)
	t.Log(r.Str)

	r = rf.Indexes(0)[2].(*Record)
	t.Log(r.Str)

	r = rf.Indexes(1)["three"].(*Record)
	t.Log(r.Str)
	t.Log(r.Arr1[1])
	t.Log(r.Arr2[2][0])
	t.Log(r.Arr3[0])
	t.Log(r.St.Name)
	t.Log(r.Maps)

	//go test -v .
	//=== RUN   Test_map
	//--- PASS: Test_map (0.00s)
	//map_test.go:41: 1
	//map_test.go:41: 2
	//map_test.go:41: 3
	//map_test.go:45: cat
	//map_test.go:48: cat
	//map_test.go:51: book
	//map_test.go:52: 6
	//map_test.go:53: 4
	//map_test.go:54: 6
	//map_test.go:55: name5566
	//map_test.go:56: map[Mysql1:{127.0.0.1 3306 root  db1} Mysql2:{127.0.0.2 3306 root ccc db2}]

}
