package tool

import (
	"errors"
	"fmt"
	"reflect"
)

// Contain : 判断某个元素是否在 slice,array ,map中
func Contain(target interface{}, obj interface{}) (bool, error) {
	targetVal := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		// 是否在slice/array中
		for i := 0; i < targetVal.Len(); i++ {
			if targetVal.Index(i).Interface() == obj {
				return true, nil
			}
		}
	case reflect.Map:
		// 是否在map key中
		if targetVal.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
	default:
		fmt.Println(reflect.TypeOf(target).Kind())
	}
	return false, errors.New("not in this array/slice/map")
}
