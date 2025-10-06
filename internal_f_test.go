package ijson

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestF struct{}

func (TestF) FieldName() string { return "type" }

func TestXFDecider_RegistryEntryWrongTypeError(t *testing.T) {
	ResetRegistries()

	key := typeKeyF[TestInterface, TestF, TestDiscriminator]{x: TestTypeA}
	registries[key] = "not_a_factory"

	var decider FDecider[TestInterface, TestF, TestDiscriminator]
	mx := map[string]TestDiscriminator{"type": TestTypeA}
	_, err := decider.Decide(mx)

	require.Error(t, err)
	assert.Equal(t, "registry[I: ijson.TestInterface, F: ijson.TestF, X: ijson.TestDiscriminator] entry should be func() I but is: string for X value typeA", err.Error())
}
