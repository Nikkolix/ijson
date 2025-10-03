package ijson_test

import (
	"fmt"
	"testing"

	"github.com/Nikkolix/ijson"
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

type InvalidTestStruct struct {
	Value string
}

type TestDiscriminator string

const (
	TestTypeA TestDiscriminator = "typeA"
	TestTypeB TestDiscriminator = "typeB"
)

func TestRegisterT_PointerTypeError(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[*ValidTestStruct, TestInterface, TestDiscriminator](TestTypeA)

	require.Error(t, err)
	assert.Equal(t, fmt.Sprintf("factory type %T must not be a pointer", &ValidTestStruct{}), err.Error())
}

func TestRegisterT_InterfaceNotImplementedError(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[InvalidTestStruct, TestInterface, TestDiscriminator](TestTypeA)

	require.Error(t, err)
	assert.Equal(t, "factory type ijson_test.InvalidTestStruct does not implement I type ijson_test.TestInterface", err.Error())
}

func TestRegisterT_SuccessfulRegistration(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[ValidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
	assert.NoError(t, err)
}

func TestRegister_FactoryError_NonPointer(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.Register[int, TestDiscriminator](TestTypeA, func() int {
		return 42
	})

	require.Error(t, err)
	assert.Equal(t, "factory must return a pointer type, got int", err.Error())
}

func TestRegister_DuplicateRegistrationError(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.Register[TestInterface, TestDiscriminator](TestTypeA, func() TestInterface {
		return &ValidTestStruct{Value: "test"}
	})
	assert.NoError(t, err)

	err = ijson.Register[TestInterface, TestDiscriminator](TestTypeA, func() TestInterface {
		return &ValidTestStruct{Value: "test2"}
	})

	require.Error(t, err)
	assert.Equal(t, "value typeA already registered for registry[I: ijson_test.TestInterface, X: ijson_test.TestDiscriminator]", err.Error())
}

func TestRegister_SuccessfulRegistration(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.Register[TestInterface, TestDiscriminator](TestTypeA, func() TestInterface {
		return &ValidTestStruct{Value: "test"}
	})
	assert.NoError(t, err)
}

func TestRegistryDecider_NoRegistryError(t *testing.T) {
	ijson.ResetRegistries()

	var decider ijson.RegistryDecider[TestInterface, TestDiscriminator]
	_, err := decider.Decide(TestTypeA)

	require.Error(t, err)
	assert.Equal(t, "no factory found in registry[I: ijson_test.TestInterface, X: ijson_test.TestDiscriminator] and X value typeA", err.Error())
}

func TestRegistryDecider_TypeNotRegisteredError(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[ValidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
	assert.NoError(t, err)

	var decider ijson.RegistryDecider[TestInterface, TestDiscriminator]
	_, err = decider.Decide(TestTypeB)

	require.Error(t, err)
	assert.Equal(t, "no factory found in registry[I: ijson_test.TestInterface, X: ijson_test.TestDiscriminator] and X value typeB", err.Error())
}

func TestRegistryDecider_SuccessfulDecision(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[ValidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
	assert.NoError(t, err)

	var decider ijson.RegistryDecider[TestInterface, TestDiscriminator]
	result, err := decider.Decide(TestTypeA)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.IsType(t, &ValidTestStruct{}, result)

	testStruct := result.(*ValidTestStruct)
	assert.Equal(t, "", testStruct.DoSomething())
}

func TestMultipleRegistrations(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[ValidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
	assert.NoError(t, err)

	type AnotherTestStruct struct {
		Data string
	}

	err = ijson.RegisterT[AnotherTestStruct, TestInterface, TestDiscriminator](TestTypeB)
	require.Error(t, err)
	assert.Equal(t, fmt.Sprintf("factory type %T does not implement I type ijson_test.TestInterface", AnotherTestStruct{}), err.Error())
}

func TestErrorMessageFormats(t *testing.T) {
	ijson.ResetRegistries()

	tests := []struct {
		name          string
		setupFunc     func() error
		expectedError string
	}{
		{
			name: "pointer type error",
			setupFunc: func() error {
				return ijson.RegisterT[*ValidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
			},
			expectedError: fmt.Sprintf("factory type %T must not be a pointer", &ValidTestStruct{}),
		},
		{
			name: "interface not implemented error",
			setupFunc: func() error {
				return ijson.RegisterT[InvalidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
			},
			expectedError: fmt.Sprintf("factory type %T does not implement I type ijson_test.TestInterface", InvalidTestStruct{}),
		},
		{
			name: "factory non-pointer error",
			setupFunc: func() error {
				return ijson.Register[string, TestDiscriminator](TestTypeA, func() string {
					return "test"
				})
			},
			expectedError: "factory must return a pointer type, got string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ijson.ResetRegistries()
			err := tt.setupFunc()
			require.Error(t, err)
			assert.Equal(t, tt.expectedError, err.Error())
		})
	}
}

func TestCompleteErrorFlow(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[InvalidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
	require.Error(t, err)
	assert.Equal(t, fmt.Sprintf("factory type %T does not implement I type ijson_test.TestInterface", InvalidTestStruct{}), err.Error())

	err = ijson.RegisterT[ValidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
	assert.NoError(t, err)

	err = ijson.RegisterT[ValidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
	require.Error(t, err)
	assert.Equal(t, "value typeA already registered for registry[I: ijson_test.TestInterface, X: ijson_test.TestDiscriminator]", err.Error())

	var decider ijson.RegistryDecider[TestInterface, TestDiscriminator]
	_, err = decider.Decide(TestTypeB)
	require.Error(t, err)
	assert.Equal(t, "no factory found in registry[I: ijson_test.TestInterface, X: ijson_test.TestDiscriminator] and X value typeB", err.Error())

	result, err := decider.Decide(TestTypeA)
	assert.NoError(t, err)
	assert.IsType(t, &ValidTestStruct{}, result)
}

func TestErrorsWithDifferentTypes(t *testing.T) {
	ijson.ResetRegistries()

	type AnotherInterface interface {
		AnotherMethod() int
	}

	type IntDiscriminator int

	var decider ijson.RegistryDecider[AnotherInterface, IntDiscriminator]
	_, err := decider.Decide(IntDiscriminator(1))

	require.Error(t, err)
	assert.Equal(t, "no factory found in registry[I: ijson_test.AnotherInterface, X: ijson_test.IntDiscriminator] and X value 1", err.Error())
}
