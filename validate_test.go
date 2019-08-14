package jsonbody

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	expected string
	actual   string
	numErrs  int
}{
	// keys must be present, value types match
	{
		`{"s": "", "b": false, "n": 0, "o": {}, "a": []}`,
		`{"s": "hi", "b": true, "n": 1, "o": {}, "a": []}`,
		0,
	},
	{
		`{"s": "", "b": false, "n": 0, "o": {}, "a": []}`,
		`{"1": "hi", "2": true, "3": 1, "4": {}, "5": []}`,
		5,
	},
	{
		`{"s": "", "b": false, "n": 0, "o": {}, "a": []}`,
		`{"s": true, "b": [], "n": {}, "o": "hi", "a": 1}`,
		5,
	},
	// key order doesn't matter
	{
		`{"s": "", "b": false, "n": 0, "o": {}, "a": []}`,
		`{"n": 1, "a": [], "s": "hi", "o": {}, "b": true}`,
		0,
	},
	// all value types in actual array checked
	{
		`{ "a": [ 0 ] }`,
		`{ "a": [ 0, 1, 2 ] }`,
		0,
	},
	{
		`{ "a": [ 0 ] }`,
		`{ "a": [ false, "hi", {}, 4 ] }`,
		3,
	},
	// nested objects/arrays
	{
		`{ "o": { "s": "", "a": [ true ]}}`,
		`{ "o": { "s": "hi", "a": [ false, true, true ]}}`,
		0,
	},
	{
		`{ "a": [ { "n": 1, "b": true } ] }`,
		`{ "a": [ { "n": -1, "b": false }, { "n": 5.5, "b": true }, { "n": 0, "b": false } ] }`,
		0,
	},
	{
		`{ "o": { "s": "", "a": [ true ]}}`,
		`{ "o": { "s": "hi", "a": [ "oops", true, {} ]}}`,
		2,
	},
	{
		`{ "a": [ { "n": 1, "b": true } ] }`,
		`{ "a": [ { "n": -1, "a": false }, { "c": 5.5, "b": true }, { "n": 0, "b": false } ] }`,
		2,
	},
	// empty object/array allows anything inside
	{
		`{}`,
		`{ "a": [ true, false ], "b": 0 }`,
		0,
	},
	{
		`{ "a": [] }`,
		`{ "a": [ { "b": "hi" }, true, 7 ] }`,
		0,
	},
	{
		`{ "s": "", "o": {} }`,
		`{ "s": 5, "o": { "b": true } }`,
		1,
	},
	// unextected empty actual objects still cause errors
	{
		`{"s": "", "b": false, "n": 0, "o": {}, "a": []}`,
		`{}`,
		5,
	},
	// empty arrays don't cause errors
	{
		`{"a": [ 0 ]}`,
		`{ "a": [] }`,
		0,
	},
}

func TestValidateReqBodyWorks(t *testing.T) {
	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			var expected, actual map[string]interface{}
			json.Unmarshal([]byte(test.expected), &expected)
			json.Unmarshal([]byte(test.actual), &actual)
			errs := validateReqBody(expected, actual)
			if len(errs) != test.numErrs {
				t.Errorf("got %v errs, want %v errs\ngot errs: %v", len(errs), test.numErrs, errs)
			}
		})
	}
}

func TestValidateReqBodyReturnsNoErrorsIfExpectedNil(t *testing.T) {
	errs := validateReqBody(nil, map[string]interface{}{})
	assert.Equal(t, 0, len(errs))
}

func TestValidateReqBodyReturnsErrorIfActualNil(t *testing.T) {
	errs := validateReqBody(map[string]interface{}{}, nil)
	assert.Equal(t, 1, len(errs))
}

func TestSetRequestSchemaSetsSchemaToNilIfNil(t *testing.T) {
	m := Middleware{}
	err := m.SetRequestSchema(http.MethodGet, nil)
	assert.Equal(t, nil, err)

	assert.Equal(t, map[string]interface{}(nil), m.reqSchemas[http.MethodGet])
}

func TestSetRequestSchemaSetsIfNotNil(t *testing.T) {
	m := Middleware{}
	err := m.SetRequestSchema(http.MethodPost, []byte("{}"))
	assert.Equal(t, nil, err)

	assert.NotEqual(t, nil, m.reqSchemas[http.MethodPost])
}
