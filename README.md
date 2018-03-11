# provide

A dependency injector for Golang. You can tell `provide` how to automatically
construct values by giving them a `PleaseProvide` initialization method,
or by annotating struct fields:

```go
type Foo struct {
    Bar *Bar `provide:""`
    Baz *Baz
}

func (foo *Foo) PleaseProvide(baz *Baz) {
    foo.Baz = baz
}
```

You can give a specific `Provider` additional rules to select interface implementations or construct values in a more customized way:

```go
provider, err := NewProvider(
    func(impl *MyImpl) MyInterface {
        return impl
    },
    func(a A, b B) (X, Y, error) {
        x := makeX(a, b)
        y := makeY(a, b)
        return x, y, errIfInvalid(x, y)
    },
)
```

There is even limited support for circular dependencies:

```go
type Foo struct {
    Bar *Bar `provide:"circular"`
}

type Bar struct {
    Foo *Foo `provide:""`
}
```

Once the `Provider` has been set up, you can use it to construct and initialize whatever values you need:

```go
var foo *Foo
var myInterface MyInterface
err := provider.Provide(&foo, &myInterface)
```
