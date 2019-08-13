package jsonbody

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

// SetRequestSchema parses a JSON schema representing the expected contents of
// a request body.
func (m *Middleware) SetRequestSchema(schemaJSON []byte) error {
	if schemaJSON == nil {
		m.reqSchema = nil
		return nil
	}

	err := json.Unmarshal(schemaJSON, &m.reqSchema)
	if err != nil {
		log.Println(fmt.Errorf("jsonbody: failed to decode schema: %v", err))
		return errors.New("failed to decode schema")
	}

	return nil
}

// TODO: create set schema method that parses expected json
func validateReqBody(expected map[string]interface{}, actual map[string]interface{}) []string {
	return validateObject("", expected, actual)
}

func validateObject(key string, expected map[string]interface{}, actual map[string]interface{}) []string {
	if len(expected) == 0 || len(actual) == 0 {
		return []string{}
	}

	errs := make([]string, 3)
	for expectedKey, expectedVal := range expected {
		actualVal, ok := actual[expectedKey]
		if !ok {
			errs = append(errs, fmt.Sprintf("expected key %v missing", key+"."+expectedKey))
		} else {
			errs = append(errs, validateSingle(key+"."+expectedKey, expectedVal, actualVal)...)
		}
	}

	return errs
}

func validateSingle(key string, expected interface{}, actual interface{}) []string {
	errs := make([]string, 3)
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
	if len(expected) == 0 || len(actual) == 0 {
		return []string{}
	}

	errs := make([]string, 3)

	for i, actualVal := range actual {
		errs = append(errs, validateSingle(fmt.Sprintf("%v[%v]", key, i), expected[0], actualVal)...)
	}

	return errs
}
