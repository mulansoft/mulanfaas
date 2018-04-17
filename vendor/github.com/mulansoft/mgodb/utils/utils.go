package utils

import (
	"reflect"
)

func IsNil(o interface{}) bool {
	return o == nil || !reflect.ValueOf(o).Elem().IsValid()
}
