package ijson

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestInterface und ValidTestStruct werden für die Tests benötigt
// ...existing code...
type TestInterface interface {
	DoSomething() string
}

type ValidTestStruct struct {
	Value string
}

func (t *ValidTestStruct) DoSomething() string {
	return t.Value
}

type TestDiscriminator string

const (
	TestTypeA TestDiscriminator = "typeA"
)

func TestRegister_RegistryWrongTypeError(t *testing.T) {
	ResetRegistries()

	key := typeKeyFor[TestInterface, TestDiscriminator]()
	registries[key] = "invalid_registry_type"

	err := Register[TestInterface, TestDiscriminator](TestTypeA, func() TestInterface {
		return &ValidTestStruct{Value: "test"}
	})

	assert.Error(t, err)
	assert.Equal(t, "registry for type typeA has wrong type", err.Error())
}

func TestRegistryDecider_RegistryWrongTypeError(t *testing.T) {
	ResetRegistries()

	key := typeKeyFor[TestInterface, TestDiscriminator]()
	registries[key] = "invalid_registry_type"

	var decider RegistryDecider[TestInterface, TestDiscriminator]
	_, err := decider.Decide(TestTypeA)

	assert.Error(t, err)
	assert.Equal(t, "registry for type typeA has wrong X type", err.Error())
}
