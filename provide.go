package provide

import (
	"errors"
	"reflect"
	"strings"
)

// A Provider is a dependency injector.
// Using the Provide method, it can automatically construct and initialize new values.
//
// It contains a pool of values as well as rules it can use to construct new ones.
// A Provider only ever contains at most a single value for any given type,
// and it can only have a single rule used to construct the value
// for any given type.
//
// If you need to dependency inject multiple different values of the same type,
// you will have to use a wrapper type in order to distinguish them.
//
// A Provider is no longer valid to use after it returns an
// error, such as when conflicting rules are added, or a rule returns
// an error instead of successfully constructing the required value.
//
type Provider struct {
	tasks  map[task]state
	values map[reflect.Type]reflect.Value
}

// NewProvider constructs a Provider given a list of rules to use to
// construct values from their dependencies.
//
//     p, err := NewProvider(rule1, rule2, ...)
//
// is equivalent to
//
//     p := &Provider{}
//     err := p.AddRule(rule1)
//     err = p.AddRule(rule2)
//     ...
//
// aborting if there's a non-nil error.
//
func NewProvider(provideFns ...interface{}) (*Provider, error) {
	provider := &Provider{}
	provider.init()
	for _, provideFn := range provideFns {
		if err := provider.AddRule(provideFn); err != nil {
			return nil, err
		}
	}
	return provider, nil
}

// AddRule gives a Provider a way to construct more types.
// All rules should be added before a Provider is used to provide any values.
//
// A rule is any function from dependencies to new values or an error:
//
//     provider.AddRule(func(a A, b B, c C) (X, Y, error) {
//         x := makeX(a, b, c)
//         y := makeY(a, b, c)
//         return x, y, nil
//     })
//
// The error return is optional, but any interface value that implements
// error will be considered to be an error.
//
func (p *Provider) AddRule(provideFn interface{}) error {
	p.init()

	initializers, err := customProvide(provideFn)
	if err != nil {
		return err
	}

	for _, init := range initializers {
		tasks := [...]task{
			{init.Type, false},
			{init.Type, true},
		}
		for _, t := range tasks {
			if _, ok := p.tasks[t]; ok {
				return errors.New("trying to provide the same type " + t.Type.String() + " in multiple ways")
			}
		}
		p.tasks[tasks[0]] = init.Partial
		p.tasks[tasks[1]] = init.Complete
	}
	return nil
}

// Provide, given a set of non-nil pointers, will construct, initialize,
// and set the values they point to using the rules the Provider has been given.
//
// Note that if the value you are trying to get is itself a pointer (and it often is,)
// you will have to give Provide a pointer-to-pointer:
//
//     var aPtr *A
//     var bInterface B
//     err := p.Provide(&aPtr, &b)
//     // aPtr and bInterface are now non-nil
//
func (p *Provider) Provide(ptrsToRequests ...interface{}) error {
	p.init()

	for _, ptr := range ptrsToRequests {
		vptr := reflect.ValueOf(ptr)
		if vptr.Kind() != reflect.Ptr || vptr.IsNil() {
			return errors.New("arguments to Provide must be non-nil pointers (that Provide will set)")
		}

		v := vptr.Elem()
		t := v.Type()
		if isErrorType(t) {
			return errors.New("since " + t.String() + " implements error, it is considered an error and cannot be provided")
		}

		if err := p.complete(t); err != nil {
			return err
		}

		value, ok := p.values[t]
		if !ok {
			return errors.New("should never happen: couldn't find value")
		}

		v.Set(value)
	}
	return nil
}

func (p *Provider) init() {
	if p.tasks == nil {
		p.tasks = make(map[task]state)
	}
	if p.values == nil {
		p.values = make(map[reflect.Type]reflect.Value)
	}
}

func (p *Provider) complete(typ reflect.Type) error {
	goals := []task{{typ, true}}
	for i := 0; i < len(goals); i++ {
		newlyDone, err := p.do(goals[i])
		if err != nil {
			return err
		}

		for _, done := range newlyDone {
			if !done.Complete {
				goals = append(goals, task{done.Type, true})
			}
		}
	}
	return nil
}

func (p *Provider) do(t task) ([]task, error) {
	stack := []task{t}
	newlyDone := make([]task, 0, 2)
	for len(stack) > 0 {
		t = stack[len(stack)-1]
		s, err := p.state(t)
		if err != nil {
			return nil, err
		}

		if !s.Done && !s.InProgress {
			// The task's dependencies need to be scheduled.
			hasDeps := false
			for _, dep := range s.DependsOn {
				depState, err := p.state(dep)
				if err != nil {
					return nil, err
				}

				if depState.Done {
					continue
				}
				if depState.InProgress {
					i := 0
					for ; i < len(stack); i++ {
						if stack[i] == t {
							break
						}
					}

					cycle := make([]string, 0, len(stack)-i)
					for ; i < len(stack); i++ {
						cycle = append(cycle, stack[i].Type.String())
					}

					return nil, errors.New("cycle: " + strings.Join(cycle, " --> "))
				}
				stack = append(stack, dep)
				hasDeps = true
			}

			s.InProgress = true
			p.tasks[t] = s
			if hasDeps {
				continue
			}
		}

		if !s.Done && s.InProgress {
			// We're returning after dependencies have been completed.
			if s.Do != nil {
				if err = s.Do(p.values); err != nil {
					return nil, err
				}
				newlyDone = append(newlyDone, t)
			}
			s.Done = true
			s.DependsOn = nil
			s.Do = nil
			p.tasks[t] = s
		}

		// This task has been completed.
		stack = stack[:len(stack)-1]
	}

	return newlyDone, nil
}

func (p *Provider) state(t task) (state, error) {
	if s, ok := p.tasks[t]; ok {
		return s, nil
	}

	init, err := autoProvide(t.Type)
	if err != nil {
		return state{}, err
	}

	p.tasks[task{t.Type, false}] = init.Partial
	p.tasks[task{t.Type, true}] = init.Complete
	return p.tasks[t], nil
}
