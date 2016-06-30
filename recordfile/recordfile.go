package recordfile

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

var Comma = '\t'
var Comment = '#'

type Index map[interface{}]interface{}

type RecordFile struct {
	Comma      rune
	Comment    rune
	typeRecord reflect.Type
	records    []interface{}
	indexes    []Index
}

func New(st interface{}) (*RecordFile, error) {
	typeRecord := reflect.TypeOf(st)
	if typeRecord == nil || typeRecord.Kind() != reflect.Struct {
		return nil, errors.New("st must be a struct")
	}

	for i := 0; i < typeRecord.NumField(); i++ {
		f := typeRecord.Field(i)

		kind := f.Type.Kind()
		switch kind {
		case reflect.Bool:
		case reflect.Int:
		case reflect.Int8:
		case reflect.Int16:
		case reflect.Int32:
		case reflect.Int64:
		case reflect.Uint:
		case reflect.Uint8:
		case reflect.Uint16:
		case reflect.Uint32:
		case reflect.Uint64:
		case reflect.Float32:
		case reflect.Float64:
		case reflect.String:
		case reflect.Struct:
		case reflect.Array:
		case reflect.Slice:
		case reflect.Map:
		default:
			return nil, fmt.Errorf("invalid type: %v %s",
				f.Name, kind)
		}

		tag := f.Tag
		if tag == "index" {
			switch kind {
			case reflect.Struct, reflect.Slice, reflect.Map:
				return nil, fmt.Errorf("could not index %s field %v %v",
					kind, i, f.Name)
			}
		}
	}

	rf := new(RecordFile)
	rf.typeRecord = typeRecord

	return rf, nil
}

func (rf *RecordFile) Read(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	if rf.Comma == 0 {
		rf.Comma = Comma
	}
	if rf.Comment == 0 {
		rf.Comment = Comment
	}
	reader := csv.NewReader(file)
	reader.Comma = rf.Comma
	reader.Comment = rf.Comment
	lines, err := reader.ReadAll()
	if err != nil {
		return err
	}

	typeRecord := rf.typeRecord

	// make records
	records := make([]interface{}, len(lines)-1)

	// make indexes
	indexes := []Index{}
	for i := 0; i < typeRecord.NumField(); i++ {
		tag := typeRecord.Field(i).Tag
		if tag == "index" {
			indexes = append(indexes, make(Index))
		}
	}

	for n := 1; n < len(lines); n++ {
		value := reflect.New(typeRecord)
		records[n-1] = value.Interface()
		record := value.Elem()

		line := lines[n]
		if len(line) != typeRecord.NumField() {
			return fmt.Errorf("line %v, field count mismatch: %v (file) %v (st)",
				n, len(line), typeRecord.NumField())
		}

		iIndex := 0

		for i := 0; i < typeRecord.NumField(); i++ {
			f := typeRecord.Field(i)

			// records
			strField := line[i]
			field := record.Field(i)
			if !field.CanSet() {
				continue
			}

			var err error

			kind := f.Type.Kind()
			if kind == reflect.Bool {
				var v bool
				v, err = strconv.ParseBool(strField)
				if err == nil {
					field.SetBool(v)
				}
			} else if kind == reflect.Int ||
				kind == reflect.Int8 ||
				kind == reflect.Int16 ||
				kind == reflect.Int32 ||
				kind == reflect.Int64 {
				var v int64
				v, err = strconv.ParseInt(strField, 0, f.Type.Bits())
				if err == nil {
					field.SetInt(v)
				}
			} else if kind == reflect.Uint ||
				kind == reflect.Uint8 ||
				kind == reflect.Uint16 ||
				kind == reflect.Uint32 ||
				kind == reflect.Uint64 {
				var v uint64
				v, err = strconv.ParseUint(strField, 0, f.Type.Bits())
				if err == nil {
					field.SetUint(v)
				}
			} else if kind == reflect.Float32 ||
				kind == reflect.Float64 {
				var v float64
				v, err = strconv.ParseFloat(strField, f.Type.Bits())
				if err == nil {
					field.SetFloat(v)
				}
			} else if kind == reflect.String {
				field.SetString(strField)
			} else if kind == reflect.Struct ||
				kind == reflect.Array ||
				kind == reflect.Slice ||
				kind == reflect.Map {
				err = json.Unmarshal([]byte(strField), field.Addr().Interface())
			}

			if err != nil {
				return fmt.Errorf("parse field (row=%v, col=%v) error: %v",
					n, i, err)
			}

			// indexes
			if f.Tag == "index" {
				index := indexes[iIndex]
				iIndex++
				if _, ok := index[field.Interface()]; ok {
					return fmt.Errorf("index error: duplicate at (row=%v, col=%v)",
						n, i)
				}
				index[field.Interface()] = records[n-1]
			}
		}
	}

	rf.records = records
	rf.indexes = indexes

	return nil
}

func (rf *RecordFile) Record(i int) interface{} {
	return rf.records[i]
}

func (rf *RecordFile) NumRecord() int {
	return len(rf.records)
}

func (rf *RecordFile) Indexes(i int) Index {
	if i >= len(rf.indexes) {
		return nil
	}
	return rf.indexes[i]
}

func (rf *RecordFile) Index(i interface{}) interface{} {
	index := rf.Indexes(0)
	if index == nil {
		return nil
	}
	return index[i]
}
