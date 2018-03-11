package provide

import "reflect"

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func isErrorType(typ reflect.Type) bool {
	return typ.Kind() == reflect.Interface && typ.Implements(errorType)
}

func isReferenceType(typ reflect.Type) bool {
	k := typ.Kind()
	return k == reflect.Ptr || k == reflect.Interface || k == reflect.Map || k == reflect.Chan
}
