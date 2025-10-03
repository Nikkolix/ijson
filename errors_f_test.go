package ijson_test

import (
	"testing"

	"github.com/Nikkolix/ijson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestFSelector struct{}

func (TestFSelector) FieldName() string { return "type" }

func TestRegisterXF_FactoryError_NonPointer(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterF[int, TestFSelector]("A", func() int { return 1 })

	require.Error(t, err)
	assert.Equal(t, "factory must return a pointer type, got int", err.Error())
}

func TestRegisterXF_DuplicateRegistrationError(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterF[TestInterface, TestFSelector]("A", func() TestInterface {
		return &ValidTestStruct{Value: "v1"}
	})
	assert.NoError(t, err)

	err = ijson.RegisterF[TestInterface, TestFSelector]("A", func() TestInterface {
		return &ValidTestStruct{Value: "v1"}
	})

	require.Error(t, err)
	assert.Equal(t, "value A already registered for registry[I: ijson_test.TestInterface, F: ijson_test.TestFSelector, X: string]", err.Error())
}

func TestRegisterXF_SuccessfulRegistration(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterF[TestInterface, TestFSelector]("A", func() TestInterface {
		return &ValidTestStruct{Value: "ok"}
	})
	assert.NoError(t, err)
}

func TestResetRegistriesXF_ClearsEntries(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterF[TestInterface, TestFSelector]("A", func() TestInterface { return &ValidTestStruct{} })
	assert.NoError(t, err)

	ijson.ResetRegistries()

	err = ijson.RegisterF[TestInterface, TestFSelector]("A", func() TestInterface { return &ValidTestStruct{} })
	assert.NoError(t, err)
}
