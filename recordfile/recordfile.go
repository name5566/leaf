package recordfile

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

var Comma rune = ','
var Comment rune = '#'

type RecordFile struct {
	Comma      rune
	Comment    rune
	kindFields []reflect.Kind
	tagFields  []reflect.StructTag
	typeRecord reflect.Type
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
	rf.typeRecord = t

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		k := f.Type.Kind()
		if k == reflect.Bool ||
			k == reflect.Int8 ||
			k == reflect.Int16 ||
			k == reflect.Int32 ||
			k == reflect.Int64 ||
			k == reflect.Uint8 ||
			k == reflect.Uint16 ||
			k == reflect.Uint32 ||
			k == reflect.Uint64 ||
			k == reflect.Float32 ||
			k == reflect.Float64 ||
			k == reflect.String {
			rf.kindFields[i] = k
		} else {
			return nil, errors.New("invalid type: " + k.String())
		}

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
		rf.Comma = Comma
	}
	if rf.Comment == 0 {
		rf.Comment = Comment
	}

	reader := csv.NewReader(f)
	reader.Comma = rf.Comma
	reader.Comment = rf.Comment
	lines, err := reader.ReadAll()
	if err != nil {
		return err
	}

	records := make([]interface{}, len(lines)-1)
	for n := 1; n < len(lines); n++ {
		v := reflect.New(rf.typeRecord)
		records[n-1] = v.Interface()
		record := v.Elem()

		line := lines[n]
		if len(line) != len(rf.kindFields) {
			return errors.New(fmt.Sprintf("line %v, field count mismatch: %v %v",
				n, len(line), len(rf.kindFields)))
		}

		for i, k := range rf.kindFields {
			fieldStr := line[i]
			field := record.Field(i)

			var err error

			if k == reflect.Bool {
				var v bool
				v, err = strconv.ParseBool(fieldStr)
				if err == nil {
					field.SetBool(v)
				}
			} else if k == reflect.Int8 ||
				k == reflect.Int16 ||
				k == reflect.Int32 ||
				k == reflect.Int64 {
				var v int64
				v, err = strconv.ParseInt(fieldStr, 0, field.Type().Bits())
				if err == nil {
					field.SetInt(v)
				}
			} else if k == reflect.Uint8 ||
				k == reflect.Uint16 ||
				k == reflect.Uint32 ||
				k == reflect.Uint64 {
				var v uint64
				v, err = strconv.ParseUint(fieldStr, 0, field.Type().Bits())
				if err == nil {
					field.SetUint(v)
				}
			} else if k == reflect.Float32 ||
				k == reflect.Float64 {
				var v float64
				v, err = strconv.ParseFloat(fieldStr, field.Type().Bits())
				if err == nil {
					field.SetFloat(v)
				}
			} else if k == reflect.String {
				field.SetString(fieldStr)
			}

			if err != nil {
				return errors.New(fmt.Sprintf("parse field (row=%v, col=%v) error: %v",
					n, i, err))
			}
		}
	}

	rf.records = records

	return nil
}

func (rf *RecordFile) Record(i int) interface{} {
	return rf.records[i]
}

func (rf *RecordFile) NumRecord() int {
	return len(rf.records)
}
