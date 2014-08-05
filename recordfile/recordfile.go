package recordfile

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"reflect"
)

const (
	defComma   = ','
	defComment = '#'
)

type RecordFile struct {
	Comma      rune
	Comment    rune
	kindFields []reflect.Kind
	tagFields  []reflect.StructTag
	records    []interface{}
	indexes    [](map[interface{}]interface{})
}

func New(st interface{}) (*RecordFile, error) {
	t := reflect.TypeOf(st)
	if t == nil || t.Kind() != reflect.Struct {
		return nil, errors.New("st must be a struct")
	}

	rf := new(RecordFile)
	rf.kindFields = make([]reflect.Kind, t.NumField())
	rf.tagFields = make([]reflect.StructTag, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// kind
		k := f.Type.Kind()
		if k == reflect.Bool ||
			k == reflect.Int8 ||
			k == reflect.Int16 ||
			k == reflect.Int32 ||
			k == reflect.Int64 ||
			k == reflect.Uint8 ||
			k == reflect.Uint32 ||
			k == reflect.Uint64 ||
			k == reflect.Float32 ||
			k == reflect.Float64 ||
			k == reflect.String {
			rf.kindFields[i] = k
		} else {
			return nil, errors.New("invalid type: " + k.String())
		}

		// tag
		tag := f.Tag
		if tag == "" ||
			tag == "index" {
			rf.tagFields[i] = tag
		} else {
			return nil, errors.New("invalid tag: " + string(tag))
		}
	}

	return rf, nil
}

func (rf *RecordFile) Read(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	if rf.Comma == 0 {
		rf.Comma = defComma
	}
	if rf.Comment == 0 {
		rf.Comment = defComment
	}

	reader := csv.NewReader(f)
	reader.Comma = rf.Comma
	reader.Comment = rf.Comment
	lines, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for n, line := range lines {
		// kind
		if n == 0 {
			for i, k := range rf.kindFields {
				fmt.Println(i, k)
				fmt.Println(line[i])
			}
		}

		// desc
		if n == 1 {
			continue
		}

		// data
	}

	return nil
}
