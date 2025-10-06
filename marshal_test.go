package ijson_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/Nikkolix/ijson"
)

// Test types for marshal testing
type ComplexStruct struct {
	Name     string            `json:"name"`
	Age      int               `json:"age"`
	Active   bool              `json:"active"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

func (c *ComplexStruct) DoSomething() string {
	return c.Name
}

type ComplexMsgpackStruct struct {
	Name     string            `msgpack:"name"`
	Age      int               `msgpack:"age"`
	Active   bool              `msgpack:"active"`
	Tags     []string          `msgpack:"tags"`
	Metadata map[string]string `msgpack:"metadata"`
}

func (c *ComplexMsgpackStruct) DoSomething() string {
	return c.Name
}

type SelfDecidingStruct struct {
	Data string `json:"data"`
}

func (s SelfDecidingStruct) DoSomething() string {
	return s.Data
}

func (s SelfDecidingStruct) Decide() (TestInterface, error) {
	return &SelfDecidingStruct{Data: s.Data}, nil
}

type SelfDecidingMsgpackStruct struct {
	Data string `msgpack:"data"`
}

func (s SelfDecidingMsgpackStruct) DoSomething() string {
	return s.Data
}

func (s SelfDecidingMsgpackStruct) Decide() (TestInterface, error) {
	return &SelfDecidingMsgpackStruct{Data: s.Data}, nil
}

type EmptyStruct struct{}

func (e *EmptyStruct) DoSomething() string {
	return ""
}

func TestDecodable_MarshalJSON_Success(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[ValidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
	assert.NoError(t, err)

	testStruct := &ValidTestStruct{Value: "test_value"}
	decodable := ijson.RDecodable[TestInterface, TestDiscriminator]{I: testStruct}

	data, err := decodable.MarshalJSON()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	expected := `{"Value":"test_value"}`
	assert.JSONEq(t, expected, string(data))
}

func TestDecodable_MarshalJSON_NilInterface(t *testing.T) {
	decodable := ijson.RDecodable[TestInterface, TestDiscriminator]{}

	data, err := decodable.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

func TestDecodable_MarshalJSON_ComplexStruct(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[ComplexStruct, TestInterface, TestDiscriminator](TestTypeA)
	assert.NoError(t, err)

	complexData := &ComplexStruct{
		Name:   "John Doe",
		Age:    30,
		Active: true,
		Tags:   []string{"developer", "golang"},
		Metadata: map[string]string{
			"department": "engineering",
			"level":      "senior",
		},
	}

	decodable := ijson.RDecodable[TestInterface, TestDiscriminator]{I: complexData}

	data, err := decodable.MarshalJSON()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", result["name"])
	assert.Equal(t, float64(30), result["age"])
	assert.Equal(t, true, result["active"])
}

func TestDecodable_MarshalMsgpack_Success(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[ValidTestStruct, TestInterface, TestDiscriminator](TestTypeA)
	assert.NoError(t, err)

	testStruct := &ValidTestStruct{Value: "test_value"}
	decodable := ijson.RDecodable[TestInterface, TestDiscriminator]{I: testStruct}

	data, err := decodable.MarshalMsgpack()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	var result ValidTestStruct
	err = msgpack.Unmarshal(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, "test_value", result.Value)
}

func TestDecodable_MarshalMsgpack_NilInterface(t *testing.T) {
	decodable := ijson.RDecodable[TestInterface, TestDiscriminator]{}

	data, err := decodable.MarshalMsgpack()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	assert.Equal(t, []byte{0xc0}, data)
}

func TestDecodable_MarshalMsgpack_ComplexStruct(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[ComplexMsgpackStruct, TestInterface, TestDiscriminator](TestTypeA)
	assert.NoError(t, err)

	complexData := &ComplexMsgpackStruct{
		Name:   "Jane Doe",
		Age:    25,
		Active: false,
		Tags:   []string{"designer", "ui/ux"},
		Metadata: map[string]string{
			"team":     "design",
			"location": "remote",
		},
	}

	decodable := ijson.RDecodable[TestInterface, TestDiscriminator]{I: complexData}

	data, err := decodable.MarshalMsgpack()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	var result ComplexMsgpackStruct
	err = msgpack.Unmarshal(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", result.Name)
	assert.Equal(t, 25, result.Age)
	assert.Equal(t, false, result.Active)
	assert.Equal(t, []string{"designer", "ui/ux"}, result.Tags)
	assert.Equal(t, map[string]string{"team": "design", "location": "remote"}, result.Metadata)
}

func TestDecodable_MarshalJSON_WithXDecodable(t *testing.T) {
	selfDeciding := SelfDecidingStruct{Data: "self_deciding_value"}
	xDecodable := ijson.XDecodable[TestInterface, SelfDecidingStruct]{I: &selfDeciding}

	data, err := xDecodable.MarshalJSON()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	expected := `{"data":"self_deciding_value"}`
	assert.JSONEq(t, expected, string(data))
}

func TestDecodable_MarshalMsgpack_WithXDecodable(t *testing.T) {
	selfDeciding := SelfDecidingMsgpackStruct{Data: "msgpack_self_deciding"}
	xDecodable := ijson.XDecodable[TestInterface, SelfDecidingMsgpackStruct]{I: &selfDeciding}

	data, err := xDecodable.MarshalMsgpack()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	var result SelfDecidingMsgpackStruct
	err = msgpack.Unmarshal(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, "msgpack_self_deciding", result.Data)
}

func TestDecodable_Marshal_EmptyStruct(t *testing.T) {
	ijson.ResetRegistries()
	err := ijson.RegisterT[EmptyStruct, TestInterface, TestDiscriminator](TestTypeA)
	assert.NoError(t, err)

	emptyStruct := &EmptyStruct{}
	decodable := ijson.RDecodable[TestInterface, TestDiscriminator]{I: emptyStruct}

	jsonData, err := decodable.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(jsonData))

	msgpackData, err := decodable.MarshalMsgpack()
	assert.NoError(t, err)
	assert.NotNil(t, msgpackData)

	var result EmptyStruct
	err = msgpack.Unmarshal(msgpackData, &result)
	assert.NoError(t, err)
}
