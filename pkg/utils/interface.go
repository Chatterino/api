package utils

import "reflect"

func IsInterfaceNil(p any) bool {
	checkNil := reflect.ValueOf(p)
	return !checkNil.IsValid() || checkNil.IsNil()
}
