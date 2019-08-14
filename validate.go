package jsonbody

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

// SetRequestSchema parses a JSON schema representing the expected contents of
// the body of a request with the given HTTP method.
//
// The schemaJSON should essentially be a sample request body. All keys in the
// schemaJSON (including those nested in objects/arrays) will be expected to be
// present in request bodies that pass through the Middleware. Additionally, all
// values will be expected to have the same type as the values in the schema.
// Arrays in the schema need only have one element in them against which all
// array elements in the real request will be verified. Finally, an empty object
// or empty array in the schema indicates that the object/array in the requests
// must be present but can have any contents.
//
// Setting schemaJSON to "" (the empty string) indicates that any body (including
// none at all) should be accepted.
func (m *Middleware) SetRequestSchema(method string, schemaJSON string) error {
	if m.reqSchemas == nil {
		m.reqSchemas = map[string]map[string]interface{}{}
	}

	if schemaJSON == "" {
		m.reqSchemas[method] = nil
		return nil
	}

	var schema map[string]interface{}
	err := json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		log.Println(fmt.Errorf("jsonbody: failed to decode schema: %v", err))
		return errors.New("failed to decode schema")
	}

	m.reqSchemas[method] = schema

	return nil
}

func validateReqBody(expected map[string]interface{}, actual map[string]interface{}) []string {
	if expected == nil {
		return []string{}
	}

	if actual == nil {
		return []string{"expected request body"}
	}

	return validateObject("", expected, actual)
}

func validateObject(key string, expected map[string]interface{}, actual map[string]interface{}) []string {
	if len(expected) == 0 {
		return []string{}
	}

	errs := make([]string, 0)
	for expectedKey, expectedVal := range expected {
		var newKey string
		if key == "" {
			newKey = expectedKey
		} else {
			newKey = key + "." + expectedKey
		}

		actualVal, ok := actual[expectedKey]
		if !ok {
			errs = append(errs, fmt.Sprintf("expected key %v missing", newKey))
		} else {
			errs = append(errs, validateSingle(newKey, expectedVal, actualVal)...)
		}
	}

	return errs
}

func validateSingle(key string, expected interface{}, actual interface{}) []string {
	errs := make([]string, 0)
	switch expected := expected.(type) {
	case string:
		if _, ok := actual.(string); !ok {
			errs = append(errs, fmt.Sprintf("value for key %v expected to be of type string", key))
		}
	case bool:
		if _, ok := actual.(bool); !ok {
			errs = append(errs, fmt.Sprintf("value for key %v expected to be of type boolean", key))
		}
	case float64:
		if _, ok := actual.(float64); !ok {
			errs = append(errs, fmt.Sprintf("value for key %v expected to be of type number", key))
		}
	case []interface{}:
		if actualArray, ok := actual.([]interface{}); !ok {
			errs = append(errs, fmt.Sprintf("value for key %v expected to be of type array", key))
		} else {
			errs = append(errs, validateArray(key, expected, actualArray)...)
		}
	case map[string]interface{}:
		if actualObj, ok := actual.(map[string]interface{}); !ok {
			errs = append(errs, fmt.Sprintf("value for key %v expected to be of type object", key))
		} else {
			errs = append(errs, validateObject(key, expected, actualObj)...)
		}
	}

	return errs
}

func validateArray(key string, expected []interface{}, actual []interface{}) []string {
	if len(expected) == 0 {
		return []string{}
	}

	errs := make([]string, 0)

	for i, actualVal := range actual {
		errs = append(errs, validateSingle(fmt.Sprintf("%v[%v]", key, i), expected[0], actualVal)...)
	}

	return errs
}
