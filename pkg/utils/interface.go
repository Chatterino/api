package utils

import "reflect"

func IsInterfaceNil(p interface{}) bool {
	checkNil := reflect.ValueOf(p)
	return !checkNil.IsValid() || checkNil.IsNil()
}
