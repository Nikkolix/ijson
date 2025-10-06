package ijson_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/Nikkolix/ijson"
)

type I interface {
	i()
}

type SA struct {
	A    string
	Type string
}

func (SA) i() {}

type SB struct {
	B    int
	Type string
}

func (SB) i() {}

type X struct {
	Type string
}

type CustomDecider struct{}

func (CustomDecider) Decide(x X) (I, error) {
	switch x.Type {
	case "SA":
		return &SA{}, nil
	case "SB":
		return &SB{}, nil
	default:
		return nil, nil
	}
}

type XDeciderImpl struct {
	Type string
}

func (x XDeciderImpl) Decide() (I, error) {
	switch x.Type {
	case "SA":
		return &SA{}, nil
	case "SB":
		return &SB{}, nil
	default:
		return nil, nil
	}
}

type UnmarshalTestInterface interface {
	GetType() string
}

type PersonStruct struct {
	Name string `json:"name" msgpack:"name"`
	Age  int    `json:"age" msgpack:"age"`
	Type string `json:"type" msgpack:"type"`
}

func (p *PersonStruct) GetType() string {
	return p.Type
}

type AnimalStruct struct {
	Species string `json:"species" msgpack:"species"`
	Sound   string `json:"sound" msgpack:"sound"`
	Type    string `json:"type" msgpack:"type"`
}

func (a *AnimalStruct) GetType() string {
	return a.Type
}

type UnmarshalComplexStruct struct {
	Type     string                 `json:"type"`
	Data     map[string]interface{} `json:"data"`
	Tags     []string               `json:"tags"`
	Metadata map[string]string      `json:"metadata"`
}

func (c *UnmarshalComplexStruct) GetType() string {
	return c.Type
}

type UnmarshalComplexMsgpackStruct struct {
	Type    string         `msgpack:"type"`
	Numbers []int          `msgpack:"numbers"`
	Nested  map[string]int `msgpack:"nested"`
	Active  bool           `msgpack:"active"`
}

func (c *UnmarshalComplexMsgpackStruct) GetType() string {
	return c.Type
}

type SelfUnmarshalStruct struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func (s SelfUnmarshalStruct) GetType() string {
	return "self"
}

func (s SelfUnmarshalStruct) Decide() (UnmarshalTestInterface, error) {
	return &SelfUnmarshalStruct{Name: s.Name, ID: s.ID}, nil
}

type SelfMsgpackStruct struct {
	Value string `msgpack:"value"`
	Count int    `msgpack:"count"`
}

func (s SelfMsgpackStruct) GetType() string {
	return "self_msgpack"
}

func (s SelfMsgpackStruct) Decide() (UnmarshalTestInterface, error) {
	return &SelfMsgpackStruct{Value: s.Value, Count: s.Count}, nil
}

type LargeDataStruct struct {
	Type   string   `msgpack:"type"`
	Items  []string `msgpack:"items"`
	Buffer []byte   `msgpack:"buffer"`
}

func (l *LargeDataStruct) GetType() string {
	return l.Type
}

type ErrorDeciderStruct struct {
	ShouldError bool `json:"should_error" msgpack:"should_error"`
}

func (e ErrorDeciderStruct) GetType() string {
	return "error_test"
}

func (e ErrorDeciderStruct) Decide() (UnmarshalTestInterface, error) {
	if e.ShouldError {
		return nil, fmt.Errorf("intentional decider error")
	}
	return &ErrorDeciderStruct{ShouldError: false}, nil
}

type InconsistentStruct struct {
	Type string `json:"type" msgpack:"type"`
	Data string `json:"data,omitempty" msgpack:"data,omitempty"`
}

func (i *InconsistentStruct) GetType() string {
	return i.Type
}

type UnmarshalDiscriminator struct {
	Type string `json:"type" msgpack:"type"`
}

const (
	PersonType = "person"
	AnimalType = "animal"
)

func TestDecodable_UnmarshalJson(t *testing.T) {
	type S struct {
		I ijson.Decodable[I, X, CustomDecider]
	}

	s := S{}
	err := json.Unmarshal(
		[]byte(`{"I":{"A":"hello","Type":"SA"}}`),
		&s,
	)

	assert.NoError(t, err)
	assert.IsType(t, &SA{}, s.I.I)
}

func TestDecodable_UnmarshalMsgpack(t *testing.T) {
	type S struct {
		I ijson.Decodable[I, X, CustomDecider]
	}

	s := S{}
	err := msgpack.Unmarshal(
		[]byte("\x81\xa1I\x82\xa1A\xa5hello\xa4Type\xa2SA"),
		&s,
	)

	assert.NoError(t, err)
	assert.IsType(t, &SA{}, s.I.I)
}

func TestRDecodable_UnmarshalJson(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[SA, I](X{"SA"})
	assert.NoError(t, err)
	err = ijson.RegisterT[SB, I](X{"SB"})
	assert.NoError(t, err)

	type S struct {
		I ijson.RDecodable[I, X]
	}

	s := S{}
	err = json.Unmarshal(
		[]byte(`{"I":{"A":"hello","Type":"SA"}}`),
		&s,
	)

	assert.NoError(t, err)
	assert.IsType(t, &SA{}, s.I.I)
}

func TestRDecodable_UnmarshalMsgpack(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[SA, I](X{"SA"})
	assert.NoError(t, err)
	err = ijson.RegisterT[SB, I](X{"SB"})
	assert.NoError(t, err)

	type S struct {
		I ijson.RDecodable[I, X]
	}

	s := S{}
	err = msgpack.Unmarshal(
		[]byte("\x81\xa1I\x82\xa1A\xa5hello\xa4Type\xa2SA"),
		&s,
	)

	assert.NoError(t, err)
	assert.IsType(t, &SA{}, s.I.I)
}

func TestXDecodable_UnmarshalJson(t *testing.T) {
	type S struct {
		I ijson.XDecodable[I, XDeciderImpl]
	}

	s := S{}
	err := json.Unmarshal(
		[]byte(`{"I":{"A":"hello","Type":"SA"}}`),
		&s,
	)

	assert.NoError(t, err)
	assert.IsType(t, &SA{}, s.I.I)
}

func TestXDecodable_UnmarshalMsgpack(t *testing.T) {
	type S struct {
		I ijson.XDecodable[I, XDeciderImpl]
	}

	s := S{}
	err := msgpack.Unmarshal(
		[]byte("\x81\xa1I\x82\xa1A\xa5hello\xa4Type\xa2SA"),
		&s,
	)

	assert.NoError(t, err)
	assert.IsType(t, &SA{}, s.I.I)
}

func TestDecodable_UnmarshalJSON_Success(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[PersonStruct, UnmarshalTestInterface, UnmarshalDiscriminator](UnmarshalDiscriminator{Type: PersonType})
	assert.NoError(t, err)
	err = ijson.RegisterT[AnimalStruct, UnmarshalTestInterface, UnmarshalDiscriminator](UnmarshalDiscriminator{Type: AnimalType})
	assert.NoError(t, err)

	personJSON := `{"name": "John Doe", "age": 30, "type": "person"}`
	var personDecodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]

	err = personDecodable.UnmarshalJSON([]byte(personJSON))
	assert.NoError(t, err)
	assert.NotNil(t, personDecodable.I)

	person, ok := personDecodable.I.(*PersonStruct)
	assert.True(t, ok)
	assert.Equal(t, "John Doe", person.Name)
	assert.Equal(t, 30, person.Age)
	assert.Equal(t, "person", person.Type)

	animalJSON := `{"species": "Dog", "sound": "Woof", "type": "animal"}`
	var animalDecodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]

	err = animalDecodable.UnmarshalJSON([]byte(animalJSON))
	assert.NoError(t, err)
	assert.NotNil(t, animalDecodable.I)

	animal, ok := animalDecodable.I.(*AnimalStruct)
	assert.True(t, ok)
	assert.Equal(t, "Dog", animal.Species)
	assert.Equal(t, "Woof", animal.Sound)
	assert.Equal(t, "animal", animal.Type)
}

func TestDecodable_UnmarshalJSON_InvalidJSON(t *testing.T) {
	ijson.ResetRegistries()

	var decodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	invalidJSON := `{"name": "John", "age": invalid}`

	err := decodable.UnmarshalJSON([]byte(invalidJSON))
	require.Error(t, err)
	assert.Equal(t, "invalid character 'i' looking for beginning of value", err.Error())
}

func TestDecodable_UnmarshalJSON_NoRegisteredType(t *testing.T) {
	ijson.ResetRegistries()

	var decodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	jsonData := `{"type": "unknown"}`

	err := decodable.UnmarshalJSON([]byte(jsonData))
	require.Error(t, err)
	assert.Equal(t, "no factory found in registry[I: ijson_test.UnmarshalTestInterface, X: ijson_test.UnmarshalDiscriminator] and X value {unknown}", err.Error())
}

func TestDecodable_UnmarshalJSON_NoRegistry(t *testing.T) {
	ijson.ResetRegistries()

	type UnknownInterface interface {
		Unknown() string
	}
	type UnknownDiscriminator struct {
		Type string `json:"type"`
	}

	var decodable ijson.RDecodable[UnknownInterface, UnknownDiscriminator]
	jsonData := `{"type": "test"}`

	err := decodable.UnmarshalJSON([]byte(jsonData))
	require.Error(t, err)
	assert.Equal(t, "no factory found in registry[I: ijson_test.UnknownInterface, X: ijson_test.UnknownDiscriminator] and X value {test}", err.Error())
}

func TestDecodable_UnmarshalJSON_ComplexStructure(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[UnmarshalComplexStruct, UnmarshalTestInterface, UnmarshalDiscriminator](UnmarshalDiscriminator{Type: "complex"})
	assert.NoError(t, err)

	complexJSON := `{
		"type": "complex",
		"data": {"key1": "value1", "key2": 42},
		"tags": ["tag1", "tag2", "tag3"],
		"metadata": {"author": "test", "version": "1.0"}
	}`

	var decodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	err = decodable.UnmarshalJSON([]byte(complexJSON))
	assert.NoError(t, err)

	complexStruct, ok := decodable.I.(*UnmarshalComplexStruct)
	assert.True(t, ok)
	assert.Equal(t, "complex", complexStruct.Type)
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, complexStruct.Tags)
	assert.Equal(t, map[string]string{"author": "test", "version": "1.0"}, complexStruct.Metadata)
}

func TestDecodable_UnmarshalJSON_EmptyJSON(t *testing.T) {
	ijson.ResetRegistries()

	var decodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	err := decodable.UnmarshalJSON([]byte("{}"))

	require.Error(t, err)
	assert.Equal(t, "no factory found in registry[I: ijson_test.UnmarshalTestInterface, X: ijson_test.UnmarshalDiscriminator] and X value {}", err.Error())
}

func TestDecodable_UnmarshalJSON_WithXDecodable(t *testing.T) {
	jsonData := `{"name": "Self Deciding", "id": 123}`
	var xDecodable ijson.XDecodable[UnmarshalTestInterface, SelfUnmarshalStruct]

	err := xDecodable.UnmarshalJSON([]byte(jsonData))
	assert.NoError(t, err)

	result, ok := xDecodable.I.(*SelfUnmarshalStruct)
	assert.True(t, ok)
	assert.Equal(t, "Self Deciding", result.Name)
	assert.Equal(t, 123, result.ID)
}

func TestDecodable_UnmarshalMsgpack_Success(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[PersonStruct, UnmarshalTestInterface, UnmarshalDiscriminator](UnmarshalDiscriminator{Type: PersonType})
	assert.NoError(t, err)
	err = ijson.RegisterT[AnimalStruct, UnmarshalTestInterface, UnmarshalDiscriminator](UnmarshalDiscriminator{Type: AnimalType})
	assert.NoError(t, err)

	personData := PersonStruct{Name: "Jane Doe", Age: 25, Type: "person"}
	personMsgpack, err := msgpack.Marshal(personData)
	assert.NoError(t, err)

	var personDecodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	err = personDecodable.UnmarshalMsgpack(personMsgpack)
	assert.NoError(t, err)
	assert.NotNil(t, personDecodable.I)

	person, ok := personDecodable.I.(*PersonStruct)
	assert.True(t, ok)
	assert.Equal(t, "Jane Doe", person.Name)
	assert.Equal(t, 25, person.Age)
	assert.Equal(t, "person", person.Type)

	animalData := AnimalStruct{Species: "Cat", Sound: "Meow", Type: "animal"}
	animalMsgpack, err := msgpack.Marshal(animalData)
	assert.NoError(t, err)

	var animalDecodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	err = animalDecodable.UnmarshalMsgpack(animalMsgpack)
	assert.NoError(t, err)
	assert.NotNil(t, animalDecodable.I)

	animal, ok := animalDecodable.I.(*AnimalStruct)
	assert.True(t, ok)
	assert.Equal(t, "Cat", animal.Species)
	assert.Equal(t, "Meow", animal.Sound)
	assert.Equal(t, "animal", animal.Type)
}

func TestDecodable_UnmarshalMsgpack_InvalidData(t *testing.T) {
	ijson.ResetRegistries()

	var decodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	invalidMsgpack := []byte{0xff, 0xff, 0xff}

	err := decodable.UnmarshalMsgpack(invalidMsgpack)
	require.Error(t, err)
	assert.Equal(t, "msgpack: unexpected code=ff decoding map length", err.Error())
}

func TestDecodable_UnmarshalMsgpack_NoRegisteredType(t *testing.T) {
	ijson.ResetRegistries()

	unknownData := UnmarshalDiscriminator{Type: "unknown"}
	msgpackData, err := msgpack.Marshal(unknownData)
	assert.NoError(t, err)

	var decodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	err = decodable.UnmarshalMsgpack(msgpackData)
	require.Error(t, err)
	assert.Equal(t, "no factory found in registry[I: ijson_test.UnmarshalTestInterface, X: ijson_test.UnmarshalDiscriminator] and X value {unknown}", err.Error())
}

func TestDecodable_UnmarshalMsgpack_ComplexStructure(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[UnmarshalComplexMsgpackStruct, UnmarshalTestInterface, UnmarshalDiscriminator](UnmarshalDiscriminator{Type: "complex_msgpack"})
	assert.NoError(t, err)

	complexData := UnmarshalComplexMsgpackStruct{
		Type:    "complex_msgpack",
		Numbers: []int{1, 2, 3, 4, 5},
		Nested:  map[string]int{"one": 1, "two": 2},
		Active:  true,
	}

	msgpackData, err := msgpack.Marshal(complexData)
	assert.NoError(t, err)

	var decodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	err = decodable.UnmarshalMsgpack(msgpackData)
	assert.NoError(t, err)

	result, ok := decodable.I.(*UnmarshalComplexMsgpackStruct)
	assert.True(t, ok)
	assert.Equal(t, "complex_msgpack", result.Type)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, result.Numbers)
	assert.Equal(t, map[string]int{"one": 1, "two": 2}, result.Nested)
	assert.Equal(t, true, result.Active)
}

func TestDecodable_UnmarshalMsgpack_WithXDecodable(t *testing.T) {
	originalData := SelfMsgpackStruct{Value: "test_value", Count: 42}
	msgpackData, err := msgpack.Marshal(originalData)
	assert.NoError(t, err)

	var xDecodable ijson.XDecodable[UnmarshalTestInterface, SelfMsgpackStruct]
	err = xDecodable.UnmarshalMsgpack(msgpackData)
	assert.NoError(t, err)

	result, ok := xDecodable.I.(*SelfMsgpackStruct)
	assert.True(t, ok)
	assert.Equal(t, "test_value", result.Value)
	assert.Equal(t, 42, result.Count)
}

func TestDecodable_UnmarshalMsgpack_EmptyData(t *testing.T) {
	ijson.ResetRegistries()

	nilMsgpack := []byte{0xc0}

	var decodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	err := decodable.UnmarshalMsgpack(nilMsgpack)
	require.Error(t, err)
	assert.Equal(t, "no factory found in registry[I: ijson_test.UnmarshalTestInterface, X: ijson_test.UnmarshalDiscriminator] and X value {}", err.Error())
}

func TestDecodable_UnmarshalMsgpack_LargeData(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[LargeDataStruct, UnmarshalTestInterface, UnmarshalDiscriminator](UnmarshalDiscriminator{Type: "large"})
	assert.NoError(t, err)

	items := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = fmt.Sprintf("item_%d", i)
	}
	buffer := make([]byte, 1024)
	for i := range buffer {
		buffer[i] = byte(i % 256)
	}

	largeData := LargeDataStruct{
		Type:   "large",
		Items:  items,
		Buffer: buffer,
	}

	msgpackData, err := msgpack.Marshal(largeData)
	assert.NoError(t, err)

	var decodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	err = decodable.UnmarshalMsgpack(msgpackData)
	assert.NoError(t, err)

	result, ok := decodable.I.(*LargeDataStruct)
	assert.True(t, ok)
	assert.Equal(t, "large", result.Type)
	assert.Len(t, result.Items, 1000)
	assert.Len(t, result.Buffer, 1024)
	assert.Equal(t, "item_0", result.Items[0])
	assert.Equal(t, "item_999", result.Items[999])
}

func TestDecodable_Unmarshal_DeciderError(t *testing.T) {
	ijson.ResetRegistries()

	jsonData := `{"should_error": true}`
	var jsonDecodable ijson.XDecodable[UnmarshalTestInterface, ErrorDeciderStruct]
	err := jsonDecodable.UnmarshalJSON([]byte(jsonData))
	require.Error(t, err)
	assert.Equal(t, "intentional decider error", err.Error())

	msgpackData, err := msgpack.Marshal(ErrorDeciderStruct{ShouldError: true})
	assert.NoError(t, err)
	var msgpackDecodable ijson.XDecodable[UnmarshalTestInterface, ErrorDeciderStruct]
	err = msgpackDecodable.UnmarshalMsgpack(msgpackData)
	require.Error(t, err)
	assert.Equal(t, "intentional decider error", err.Error())
}

func TestDecodable_Unmarshal_SecondUnmarshalFails(t *testing.T) {
	ijson.ResetRegistries()

	err := ijson.RegisterT[InconsistentStruct, UnmarshalTestInterface, UnmarshalDiscriminator](UnmarshalDiscriminator{Type: "inconsistent"})
	assert.NoError(t, err)

	invalidStructureJSON := `{"type": "inconsistent", "data": {"invalid": "structure"}}`

	var decodable ijson.RDecodable[UnmarshalTestInterface, UnmarshalDiscriminator]
	err = decodable.UnmarshalJSON([]byte(invalidStructureJSON))
	require.Error(t, err)
	assert.Equal(t, "json: cannot unmarshal object into Go struct field InconsistentStruct.data of type string", err.Error())
}
