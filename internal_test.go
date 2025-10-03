package ijson

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestRegistryDecider_RegistryWrongTypeError(t *testing.T) {
	ResetRegistries()

	key := typeKey[TestInterface, TestDiscriminator]{x: TestTypeA}
	registries[key] = "invalid_registry_type"

	var decider RegistryDecider[TestInterface, TestDiscriminator]
	_, err := decider.Decide(TestTypeA)

	require.Error(t, err)
	assert.Equal(t, "registry[I: ijson.TestInterface, X: ijson.TestDiscriminator] entry should be func() I but is: string for X value typeA", err.Error())
}
