[![Go Reference](https://pkg.go.dev/badge/github.com/Nikkolix/ijson.svg)](https://pkg.go.dev/github.com/Nikkolix/ijson)
![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen)
[![Go Report Card](https://goreportcard.com/badge/github.com/Nikkolix/ijson)](https://goreportcard.com/report/github.com/Nikkolix/ijson)
[![Build](https://github.com/Nikkolix/ijson/actions/workflows/go.yml/badge.svg)](https://github.com/Nikkolix/ijson/actions)
# ijson

A tiny generic helper to (un)marshal JSON and MessagePack into interface-backed values by deciding the concrete type at runtime.

It supports two ways to decide the concrete type:
- Registry-based: You register a mapping from a discriminator value to a factory that builds the concrete implementation. Use `RDecodable` with `RegistryDecider`.
- Self-deciding (XDecidable): The incoming payload type knows how to choose the target implementation. Use `XDecidable`.

Built on top of Go generics and integrates with both `encoding/json` and `github.com/vmihailenco/msgpack/v5`.

---

## Install

The module path is derived from `go.mod`:

```bash
go get github.com/Nikkolix/ijson
```

## Go version

This repo targets Go 1.25+ (see `go.mod`).

## Quick start (registry-based)

Suppose you want to decode JSON or MessagePack into an `Animal` interface, where a discriminator field `Type` chooses the concrete type at runtime.

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/Nikkolix/ijson"
)

type Animal interface { Speak() string }

type Dog struct{ Name string }
func (d *Dog) Speak() string { return "woof: " + d.Name }

type Cat struct{ Name string }
func (c *Cat) Speak() string { return "meow: " + c.Name }

// Discriminator read during decode. The field name must match the payload.
type Disc struct{ Type string }

func main() {
    // 1) Register mapping Disc -> factory(*T implements Animal)
    ijson.ResetRegistries()
    _ = ijson.RegisterT[Dog, Animal, Disc](Disc{Type: "dog"})
    _ = ijson.RegisterT[Cat, Animal, Disc](Disc{Type: "cat"})

    // 2) Use RDecodable[Animal, Disc] to unmarshal
    var a ijson.RDecodable[Animal, Disc]

    data := []byte(`{"Name":"Fido","Type":"dog"}`)
    if err := json.Unmarshal(data, &a); err != nil {
        panic(err)
    }

    fmt.Println(a.I.Speak()) // => woof: Fido
}
```

Notes
- `RegisterT[T, I, X](x X)` requires that `*T` implements `I` (pointer receiver is fine). It also enforces that the factory creates a pointer type.
- The registry is keyed by the full value of `X` (struct or other comparable type). What you pass in `x` at registration must equal the value parsed from the payload.

### MessagePack works the same

```go
import "github.com/vmihailenco/msgpack/v5"

// ...same setup & registrations as above...
var a ijson.RDecodable[Animal, Disc]
if err := msgpack.Unmarshal(msgpackBytes, &a); err != nil {
    // handle error
}
```

## Quick start (self-deciding XDecidable)

If the input type itself knows how to pick the target implementation, implement `Decide() (I, error)` on the payload type and use `XDecidable`:

```go
type Animal interface{ Speak() string }

type Self struct {
    Kind string
    Name string
}

// Decide which concrete Animal to allocate based on Kind.
func (s Self) Decide() (Animal, error) {
    switch s.Kind {
    case "dog": return &Dog{Name: s.Name}, nil
    case "cat": return &Cat{Name: s.Name}, nil
    default:     return nil, fmt.Errorf("unknown kind: %s", s.Kind)
    }
}

// Also implement the marker constraint (~struct) via being a struct type.

// Decode
var x ijson.XDecidable[Animal, Self]
if err := json.Unmarshal([]byte(`{"Kind":"dog","Name":"Fido"}`), &x); err != nil {
    panic(err)
}
fmt.Println(x.I.Speak())
```

## API overview

Key pieces you will typically touch:
- Types
  - `type Decodable[I any, X any, D Decider[I, X]]` (generic wrapper)
  - `type RDecodable[I any, X comparable]` = registry-based alias
  - `type XDecidable[I any, X XDecider[I, X]]` = self-deciding alias
- Registry helpers
  - `func RegisterT[T any, I any, X comparable](x X) error`
  - `func Register[I any, X comparable](x X, factory func() I) error`
  - `func ResetRegistries()`
- Deciders
  - `type RegistryDecider[I any, X comparable] struct{}` (used by `RDecodable`)
  - `type XDecider[I, X any] interface { Decide() (I, error); any }` (for `XDecidable`)
- Marshal/Unmarshal integrations
  - `Decodable.MarshalJSON / UnmarshalJSON`
  - `Decodable.MarshalMsgpack / UnmarshalMsgpack`

## Error messages you may see

- "factory type %T must not be a pointer"
- "factory type %T does not implement I type %T"
- "factory must return a pointer type, got %T"
- "type %v already registered"
- "no registry for I type %T and X type %T"
- "no factory for X type %v"

These make it clear whether the issue is the registry wiring, the factory types, or the discriminator value in the payload.

## Tips and gotchas

- Registration requires the factory to return a pointer to the concrete type. `RegisterT` enforces that by checking the dynamic type.
- For registry-based decoding, your discriminator type `X` must be comparable and reflect the incoming payload fields so it can be unmarshalled first.
- `RegistryDecider` and registry maps are protected by an internal RWMutex and are safe for concurrent reads/writes (per call), but you should generally register at startup.
- `D` in `Decodable[I,X,D]` needs to be a struct type (constraint `~struct{}`), so pass a struct as the decider (which is what the aliases already do).

## Run tests

```bash
go test -v ./...
```

## License

MIT (see LICENSE if present).
