package provide

import "reflect"

type task struct {
	Type     reflect.Type
	Complete bool
}

type state struct {
	Done       bool
	InProgress bool
	DependsOn  []task
	Do         func(map[reflect.Type]reflect.Value) error
}

type initializer struct {
	Type     reflect.Type
	Partial  state
	Complete state
}
