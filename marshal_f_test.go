package ijson_test

import (
	"encoding/json"
	"testing"

	"github.com/Nikkolix/ijson"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
)

func TestDecodableXF_MarshalJSON_Success(t *testing.T) {
	x := &XA{Type: "A", Value: "abc"}
	d := ijson.DecodableF[XFTestInterface, TestFSelector, string]{I: x}

	data, err := d.MarshalJSON()
	assert.NoError(t, err)

	var got map[string]any
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err)
	assert.Equal(t, "A", got["type"])
	assert.Equal(t, "abc", got["value"])
}

func TestDecodableXF_MarshalMsgpack_Success(t *testing.T) {
	x := &XA{Type: "A", Value: "xyz"}
	d := ijson.DecodableF[XFTestInterface, TestFSelector, string]{I: x}

	b, err := d.MarshalMsgpack()
	assert.NoError(t, err)

	var out XA
	err = msgpack.Unmarshal(b, &out)
	assert.NoError(t, err)
	assert.Equal(t, "A", out.Type)
	assert.Equal(t, "xyz", out.Value)
}

func TestDecodableXF_Marshal_NilInterface(t *testing.T) {
	var d ijson.DecodableF[XFTestInterface, TestFSelector, string]

	jb, err := d.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, "null", string(jb))

	mb, err := d.MarshalMsgpack()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xc0}, mb)
}
