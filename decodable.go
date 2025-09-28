// Copyright (c) 2025 Nikkolix. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.

// Package ijson provides generic, discriminator-based polymorphic unmarshaling
// for JSON and MessagePack.
// It supports multiple strategies for type resolution
// and works with both encoding/json and vmihailenco/msgpack.
package ijson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	_ json.Marshaler   = Decodable[any, any, RegistryDecider[any, any]]{}
	_ json.Marshaler   = &Decodable[any, any, RegistryDecider[any, any]]{}
	_ json.Unmarshaler = &Decodable[any, any, RegistryDecider[any, any]]{}

	_ msgpack.Marshaler   = Decodable[any, any, RegistryDecider[any, any]]{}
	_ msgpack.Marshaler   = &Decodable[any, any, RegistryDecider[any, any]]{}
	_ msgpack.Unmarshaler = &Decodable[any, any, RegistryDecider[any, any]]{}
)

// Decider is a generic interface that determines the concrete type to instantiate
// based on a discriminator value.
// It enables polymorphic deserialization by
// mapping discriminator values of type X to instances of interface type I.
// The ~struct{} constraint ensures that only empty struct types can implement this interface.
type Decider[I, X any] interface {
	// Decide returns a new instance of I based on the discriminator value x.
	Decide(X) (I, error)
	~struct{}
}

// Decodable is a generic wrapper for polymorphic (de)serialization.
// I is the interface type, X is the discriminator type, D is the decider.
type Decodable[I any, X any, D Decider[I, X]] struct {
	I I // The decoded value implementing I
}

// MarshalMsgpack marshals the contained value using msgpack.
func (d Decodable[I, X, D]) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(d.I)
}

// MarshalJSON marshals the contained value using JSON.
func (d Decodable[I, X, D]) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.I)
}

// UnmarshalMsgpack does unmarshal data into the contained value using msgpack.
// It uses the decider to resolve the concrete type based on the discriminator.
func (d *Decodable[I, X, D]) UnmarshalMsgpack(data []byte) error {
	x := new(X)
	err := msgpack.Unmarshal(data, x)
	if err != nil {
		return err
	}

	var decider D
	d.I, err = decider.Decide(*x)
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(data, d.I)
}

// UnmarshalJSON does unmarshal data into the contained value using JSON.
// It uses the decider to resolve the concrete type based on the discriminator.
func (d *Decodable[I, X, D]) UnmarshalJSON(data []byte) error {
	x := new(X)
	err := json.Unmarshal(data, x)
	if err != nil {
		return err
	}

	var decider D
	d.I, err = decider.Decide(*x)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, d.I)
}

// xAdapter adapts XDecider to Decider for generic use.
type xAdapter[I any, X XDecider[I, X]] struct{}

func (xAdapter[I, X]) Decide(x X) (I, error) {
	return x.Decide()
}

// XDecider is an interface for types that can decide their own concrete type.
type XDecider[I, X any] interface {
	Decide() (I, error)
	any
}

// XDecodable is a type alias for Decodable using xAdapter.
type XDecodable[I any, X XDecider[I, X]] = Decodable[I, X, xAdapter[I, X]]

// RDecodable is a type alias for Decodable using RegistryDecider.
type RDecodable[I any, X comparable] = Decodable[I, X, RegistryDecider[I, X]]

// registry is a map from discriminator value to factory function.
type registry[I any, X comparable] = map[X]func() I

type typeKey struct {
	I reflect.Type
	X reflect.Type
}

// typeKeyFor returns a unique key for registry lookup based on types.
func typeKeyFor[I any, X any]() typeKey {
	return typeKey{
		I: reflect.TypeFor[I](),
		X: reflect.TypeFor[X](),
	}
}

var registries = map[typeKey]any{}
var mutex = sync.RWMutex{}

// ResetRegistries clears all registered types. Useful for tests.
func ResetRegistries() {
	mutex.Lock()
	defer mutex.Unlock()
	clear(registries)
}

// RegisterT registers a type T for interface I and discriminator X.
// T must not be a pointer and must implement I.
func RegisterT[T any, I any, X comparable](x X) error {
	if reflect.TypeFor[T]().Kind() == reflect.Pointer {
		return fmt.Errorf("factory type %T must not be a pointer", *new(T))
	}

	if _, ok := any(new(T)).(I); !ok {
		return fmt.Errorf("factory type %T does not implement I type %T", *new(T), *new(I))
	}
	return Register[I, X](x, func() I {
		return any(new(T)).(I)
	})
}

// Register registers a factory function for interface I and discriminator X.
// The factory must return a pointer type.
func Register[I any, X comparable](x X, factory func() I) error {
	mutex.Lock()
	defer mutex.Unlock()

	t := factory()
	if reflect.TypeOf(t).Kind() != reflect.Pointer {
		return fmt.Errorf("factory must return a pointer type, got %T", t)
	}

	key := typeKeyFor[I, X]()
	genericReg, ok := registries[key]
	if !ok {
		genericReg = registry[I, X]{}
		registries[key] = genericReg
	}

	reg, ok := genericReg.(registry[I, X])
	if !ok {
		return fmt.Errorf("registry for type %v has wrong type", x)
	}

	_, ok = reg[x]
	if ok {
		return fmt.Errorf("type %v already registered", x)
	}

	reg[x] = factory
	return nil
}

// RegistryDecider resolves a concrete type from a registry based on discriminator value.
type RegistryDecider[I any, X comparable] struct {
}

// Decide returns a new instance of I from the registry for discriminator x.
func (RegistryDecider[I, X]) Decide(x X) (I, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	var i I
	genericReg, ok := registries[typeKeyFor[I, X]()]
	if !ok {
		return i, fmt.Errorf("no registry for I type %T and X type %T", i, x)
	}

	reg, ok := genericReg.(registry[I, X])
	if !ok {
		return i, fmt.Errorf("registry for type %v has wrong X type", x)
	}

	factory, ok := reg[x]
	if !ok {
		return i, fmt.Errorf("no factory for X type %v", x)
	}

	return factory(), nil
}
