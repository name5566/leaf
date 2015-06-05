package util

// reference: https://github.com/mohae/deepcopy
import (
	"reflect"
)

func DeepClone(v interface{}) interface{} {
	dst := reflect.New(reflect.TypeOf(v)).Elem()
	deepCopy(dst, reflect.ValueOf(v))
	return dst.Interface()
}

func deepCopy(dst, src reflect.Value) {
	switch src.Kind() {
	case reflect.Interface:
		deepCopy(dst.Elem(), src.Elem())
	case reflect.Ptr:
		value := src.Elem()
		if !value.IsValid() {
			return
		}
		dst.Set(reflect.New(value.Type()))
		deepCopy(dst.Elem(), value)
	case reflect.Map:
		dst.Set(reflect.MakeMap(src.Type()))
		keys := src.MapKeys()
		for _, key := range keys {
			value := src.MapIndex(key)
			newValue := reflect.New(value.Type()).Elem()
			deepCopy(newValue, value)
			dst.SetMapIndex(key, newValue)
		}
	case reflect.Slice:
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			deepCopy(dst.Index(i), src.Index(i))
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			deepCopy(dst.Field(i), src.Field(i))
		}
	default:
		dst.Set(src)
	}
}
