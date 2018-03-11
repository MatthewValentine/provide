// Package provide implements a simple dependency injector.
//
// Dependency injection allows you to automatically initialize values
// by specifying what you dependencies you need to initialize them
// and how to get those dependencies. The type that handles dependency injection
// is the Provider.
//
// Rules
//
// Rules describing how to construct new values from dependencies are
// given to a Provider as functions from dependencies to new values:
//
//     provider.AddRule(func(a A, b B, c C) (X, Y, error) {
//         x := makeX(a, b, c)
//         y := makeY(a, b, c)
//         return x, y, errIfInvalid(x, y)
//     })
//
// When either an X or Y is first needed, this rule will be automatically
// called with fully initialized values of A, B, and C.
// If the rule returns an error, the Provider will stop and return that error.
// Since a Provider only keeps around a single value of any type, a rule
// will only ever be used at most once.
//
// Interfaces
//
// Interfaces cannot be automatically constructed.
// But if you want to use an automatically constructable implementation,
// you can use a very simple rule:
//
//     func(impl *MyImplementation) MyInterface {
//         return impl
//     }
//
// The implementation will be automatically constructed and then passed
// to the rule, which basically just selects it as being the canonical
// MyInterface implementation. Being able to easily swap implementations
// like this is one of the main draws of dependency injection.
//
// Automatic rules
//
// Rather than directly adding a rule to a Provider, you can also have
// the type specify its own initialization requirements.
// You can do this by giving the type a PleaseProvide method, annotating
// the fields of structs with a `provide:""` tag, or both:
//
//     type Foo struct {
//         Bar Bar `provide:""`
//         Baz Baz
//     }
//
//     func (foo *Foo) PleaseProvide(baz Baz) error {
//         foo.Baz = baz
//         return errIfInvalid(foo)
//     }
//
// The PleaseProvide method, if any, will be called immediately after the
// annotated fields are set, if any. If a type has neither provide-annotated
// fields nor a PleaseProvide method, it will not be automatically constructed.
// If you want Providers to automatically construct your type but it doesn't
// actually have any dependencies, simply add an empty PleaseProvide method.
//
// If you give a Provider a rule that outputs an automatically-constructable type,
// the rule will take precedence and the automatic construction will not occur.
// In that case, it is up to the rule to make sure the value has been properly initialized.
//
// Circular dependencies
//
// Normally, circular dependencies are an error.
// However, by using the `provide:"circular"` annotation on a field,
// it is possible to initialize mutually dependent types in a limited way:
//
//     type Foo struct {
//         Bar *Bar `provide:"circular"`
//     }
//
//     type Bar struct {
//         Foo *Foo `provide:""`
//     }
//
// Circular fields are dangerous. The field will be set before
// its value has been fully initialized. Although they are guaranteed
// to have been fully initialized by the end of the call to the Provide
// method, there are no further guarantees. Because the value is modified
// further after being set, only pointers, maps, and channel fields can
// be circular fields.
//
// Multiple values of the same type
//
// Since provide can only tell what value a Go function is looking for
// based on its type, Providers aren't able provide multiple different values
// for the same type. If you need that functionality, you will have to
// use a wrapper type to distinguish them.
//
package provide
