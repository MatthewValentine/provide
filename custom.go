package provide

import (
	"errors"
	"reflect"
)

func customProvide(provideFn interface{}) ([]initializer, error) {
	v := reflect.ValueOf(provideFn)
	t := v.Type()

	if t.Kind() != reflect.Func {
		return nil, errors.New("providers must be functions")
	}

	ins := make([]reflect.Type, t.NumIn())
	deps := make([]task, len(ins))
	for i := range ins {
		ins[i] = t.In(i)
		deps[i] = task{ins[i], true}
	}

	type output struct {
		Type  reflect.Type
		IsErr bool
	}

	outs := make([]output, t.NumOut())
	for i := range outs {
		out := t.Out(i)
		outs[i] = output{
			Type:  out,
			IsErr: isErrorType(out),
		}
	}

	alreadyDone := false
	doFn := func(values map[reflect.Type]reflect.Value) error {
		if alreadyDone {
			return nil
		}
		alreadyDone = true

		inputs := make([]reflect.Value, len(ins))
		for i := range ins {
			inputs[i] = values[ins[i]]
		}

		outputs := v.Call(inputs)
		for i := range outputs {
			if outs[i].IsErr {
				if !outputs[i].IsNil() {
					return outputs[i].Interface().(error)
				}
			} else {
				values[outs[i].Type] = outputs[i]
			}
		}
		return nil
	}

	initializers := make([]initializer, 0, len(outs))
	for i := range outs {
		if outs[i].IsErr {
			continue
		}

		initializers = append(initializers, initializer{
			Type:     outs[i].Type,
			Partial:  state{DependsOn: deps, Do: doFn},
			Complete: state{DependsOn: []task{{outs[i].Type, false}}},
		})
	}
	return initializers, nil
}
