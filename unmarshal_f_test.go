package ijson_test

import (
	"testing"

	"github.com/Nikkolix/ijson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

type XFTestInterface interface {
	Kind() string
}

type XA struct {
	Type  string `json:"type" msgpack:"type"`
	Value string `json:"value" msgpack:"value"`
}

func (x *XA) Kind() string { return "A" }

type XB struct {
	Type  string `json:"type" msgpack:"type"`
	Value string `json:"value" msgpack:"value"`
}

func (x *XB) Kind() string { return "B" }

func TestDecodableXF_UnmarshalJSON_Success(t *testing.T) {
	ijson.ResetRegistries()

	assert.NoError(t, ijson.RegisterF[XFTestInterface, TestFSelector]("A", func() XFTestInterface { return &XA{} }))
	assert.NoError(t, ijson.RegisterF[XFTestInterface, TestFSelector]("B", func() XFTestInterface { return &XB{} }))

	jsonData := `{"type":"A","value":"hello"}`
	var d ijson.DecodableF[XFTestInterface, TestFSelector, string]
	err := d.UnmarshalJSON([]byte(jsonData))

	assert.NoError(t, err)
	if assert.IsType(t, &XA{}, d.I) {
		a := d.I.(*XA)
		assert.Equal(t, "A", a.Type)
		assert.Equal(t, "hello", a.Value)
	}
}

func TestDecodableXF_UnmarshalMsgpack_Success(t *testing.T) {
	ijson.ResetRegistries()

	assert.NoError(t, ijson.RegisterF[XFTestInterface, TestFSelector]("A", func() XFTestInterface { return &XA{} }))

	payload := map[string]string{"type": "A", "value": "v"}
	raw, err := msgpack.Marshal(payload)
	assert.NoError(t, err)

	var d ijson.DecodableF[XFTestInterface, TestFSelector, string]
	err = d.UnmarshalMsgpack(raw)

	assert.NoError(t, err)
	assert.IsType(t, &XA{}, d.I)
	assert.Equal(t, "v", d.I.(*XA).Value)
}

func TestDecodableXF_UnmarshalJSON_MissingDiscriminator(t *testing.T) {
	ijson.ResetRegistries()

	assert.NoError(t, ijson.RegisterF[XFTestInterface, TestFSelector]("A", func() XFTestInterface { return &XA{} }))

	var d ijson.DecodableF[XFTestInterface, TestFSelector, string]
	err := d.UnmarshalJSON([]byte(`{"value":"x"}`))

	require.Error(t, err)
	assert.Equal(t, "discriminator field type not found in map map[value:x]", err.Error())
}

func TestDecodableXF_UnmarshalJSON_NoRegistryEntry(t *testing.T) {
	ijson.ResetRegistries()

	var d ijson.DecodableF[XFTestInterface, TestFSelector, string]
	err := d.UnmarshalJSON([]byte(`{"type":"Z","value":"x"}`))

	require.Error(t, err)
	assert.Equal(t, "no factory found in registry[I: ijson_test.XFTestInterface, F: ijson_test.TestFSelector, X: string] and X value Z", err.Error())
}
