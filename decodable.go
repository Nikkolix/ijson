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

type Decider[I, X any] interface {
	Decide(X) (I, error)
	~struct{}
}

type Decodable[I any, X any, D Decider[I, X]] struct {
	I I
}

// marshal

func (d Decodable[I, X, D]) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(d.I)
}

func (d Decodable[I, X, D]) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.I)
}

// unmarshal

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

// x decodable

type xAdapter[I any, X XDecider[I, X]] struct{}

func (xAdapter[I, X]) Decide(x X) (I, error) {
	return x.Decide()
}

type XDecider[I, X any] interface {
	Decide() (I, error)
	any
}

type XDecodable[I any, X XDecider[I, X]] = Decodable[I, X, xAdapter[I, X]]

// registry

type RDecodable[I any, X comparable] = Decodable[I, X, RegistryDecider[I, X]]

type registry[I any, X comparable] = map[X]func() I

type typeKey struct {
	I reflect.Type
	X reflect.Type
}

func typeKeyFor[I any, X any]() typeKey {
	return typeKey{
		I: reflect.TypeFor[I](),
		X: reflect.TypeFor[X](),
	}
}

var registries = map[typeKey]any{}
var mutex = sync.RWMutex{}

func ResetRegistries() {
	mutex.Lock()
	defer mutex.Unlock()
	clear(registries)
}

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

type RegistryDecider[I any, X comparable] struct {
}

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
