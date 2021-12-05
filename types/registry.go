package types

import (
	"log"
	"reflect"
)

var typeRegistry = make(map[string]reflect.Type)

func RegisterType(name string, datatype interface{}) {
	log.Println("Registering type name:", name, reflect.TypeOf(datatype))
	typeRegistry[name] = reflect.TypeOf(datatype)
}

func CreateType(name string) interface{} {
	if t := typeRegistry[name]; t != nil {
		i := reflect.New(t).Interface()
		return i
	}

	return nil
}

func CreateSlice(name string) interface{} {
	tr := typeRegistry[name]
	if tr == nil {
		return nil
	}

	t := reflect.New(reflect.SliceOf(typeRegistry[name])).Interface()
	return t
}
