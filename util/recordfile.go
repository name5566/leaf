package recordfile

import (
	"encoding/csv"
	"errors"
	"os"
	"reflect"
)

const defComma = '\t'
const defComment = '#'

type RecordFile struct {
	Comma   rune
	Comment rune
	types   []reflect.Kind
	stType  reflect.Type
}

func New(st interface{}) (*RecordFile, error) {
	t := reflect.TypeOf(st)
	if t == nil || t.Kind() != reflect.Struct {
		return nil, errors.New("st must be a struct")
	}

	rf := new(RecordFile)
	rf.types = make([]reflect.Kind, t.NumField())
	rf.stType = t

	for i := 0; i < t.NumField(); i++ {
		k := t.Field(i).Type.Kind()
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
			rf.types[i] = k
		} else {
			return nil, errors.New("invalid type: " + k.String())
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
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// types
	if len(records) >= 1 {
		for i, v := range rf.types {

		}
	}

	if len(records) >= 3 {

	}

	return nil
}

func (rf *RecordFile) NumRecord() int {
	return len(rf.types)
}

/*
func (rf *RecordFile) Record(i int32) interface{} {
	return nil
}

func (rf *RecordFile) RecordByIndex(index int32) interface{} {
	return nil
}
*/
