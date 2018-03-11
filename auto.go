package provide

import (
	"errors"
	"reflect"
)

func autoProvide(typ reflect.Type) (initializer, error) {
	switch typ.Kind() {
	case reflect.Interface:
		return initializer{}, errors.New(typ.String() + " can't be automatically provided")

	case reflect.Ptr:
		type Field struct {
			Type           reflect.Type
			Index          int
			MustBeComplete bool
		}

		var providedFields []Field
		elem := typ.Elem()
		if elem.Kind() == reflect.Struct {
			N := elem.NumField()
			for i := 0; i < N; i++ {
				field := elem.Field(i)
				tag, ok := field.Tag.Lookup("provide")
				if !ok {
					continue
				}

				allowCircular := (tag == "circular")
				if !allowCircular {
					if tag != "" {
						return initializer{}, errors.New(
							"unrecognized provide tag " + tag + " in " + elem.String(),
						)
					}
				} else {
					if !isReferenceType(field.Type) {
						return initializer{}, errors.New(
							"only reference types (pointer, map, chan) can be circularly provided to " + elem.String() + ", not " + field.Type.String(),
						)
					}
				}

				providedFields = append(providedFields, Field{
					Type:           field.Type,
					Index:          i,
					MustBeComplete: !allowCircular,
				})
			}
		}

		initFn, hasInitFn := typ.MethodByName("PleaseProvide")
		var ins []reflect.Type
		if hasInitFn {
			ins = make([]reflect.Type, initFn.Type.NumIn()-1)
			for i := range ins {
				ins[i] = initFn.Type.In(i + 1)
			}

			nOut := initFn.Type.NumOut()
			for i := 0; i < nOut; i++ {
				if !isErrorType(initFn.Type.Out(i)) {
					return initializer{}, errors.New(
						typ.String() + ".PleaseProvide must only return errors (which must be interfaces), not " + initFn.Type.Out(i).String(),
					)
				}
			}
		}

		if len(providedFields) == 0 && !hasInitFn {
			return initializer{}, errors.New(typ.String() + " can't be automatically provided")
		}

		deps := make([]task, 0, 1+len(providedFields)+len(ins))
		deps = append(deps, task{typ, false})
		for _, field := range providedFields {
			deps = append(deps, task{field.Type, field.MustBeComplete})
		}
		for _, in := range ins {
			deps = append(deps, task{in, true})
		}

		initFnVal := initFn.Func
		doFn := func(values map[reflect.Type]reflect.Value) error {
			v := values[typ]

			if len(providedFields) > 0 {
				elem := v.Elem()
				for _, field := range providedFields {
					elem.Field(field.Index).Set(values[field.Type])
				}
			}

			if hasInitFn {
				inputs := make([]reflect.Value, len(ins)+1)
				inputs[0] = v
				for i, in := range ins {
					inputs[i+1] = values[in]
				}
				outputs := initFnVal.Call(inputs)
				for _, out := range outputs {
					if !out.IsNil() {
						return out.Interface().(error)
					}
				}
			}

			return nil
		}

		return initializer{
			Type: typ,
			Partial: state{
				Do: func(values map[reflect.Type]reflect.Value) error {
					values[typ] = reflect.New(elem)
					return nil
				},
			},
			Complete: state{
				DependsOn: deps,
				Do:        doFn,
			},
		}, nil

	default:
		ptrTo := reflect.PtrTo(typ)
		return initializer{
			Type: typ,
			Partial: state{
				DependsOn: []task{{ptrTo, true}},
				Do: func(values map[reflect.Type]reflect.Value) error {
					vptr := values[ptrTo]
					if vptr.IsNil() {
						return errors.New("can't use nil pointer to automatically provide value for " + typ.String())
					}

					values[typ] = vptr.Elem()
					return nil
				},
			},
			Complete: state{
				DependsOn: []task{{typ, false}},
			},
		}, nil
	}
}
