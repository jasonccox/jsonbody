package jsonbody

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

func parseSchema(schemaJSON string) (map[string]interface{}, error) {
	if schemaJSON == "" {
		return nil, nil
	}

	var schemaMap map[string]interface{}
	err := json.Unmarshal([]byte(schemaJSON), &schemaMap)
	if err != nil {
		log.Printf("jsonbody: failed to decode schema: %v\n", err)
		return nil, errors.New("jsonbody: failed to decode schema")
	}

	return schemaMap, nil
}

func validateReqBody(expected map[string]interface{}, actual map[string]interface{}) []string {
	if expected == nil {
		return []string{}
	}

	if actual == nil {
		return []string{"expected a JSON body"}
	}

	return validateObject("", expected, actual)
}

func validateObject(key string, expected map[string]interface{}, actual map[string]interface{}) []string {
	if len(expected) == 0 {
		return []string{}
	}

	errs := make([]string, 0)
	for expectedKey, expectedVal := range expected {
		optional := strings.HasPrefix(expectedKey, "?")
		expectedKey = strings.TrimPrefix(expectedKey, "?")

		var newKey string
		if key == "" {
			newKey = expectedKey
		} else {
			newKey = key + "." + expectedKey
		}

		actualVal, ok := actual[expectedKey]
		if !optional && !ok {
			errs = append(errs, fmt.Sprintf("expected key '%v' missing", newKey))
		} else if ok {
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
			errs = append(errs, fmt.Sprintf("value for key '%v' expected to be of type string", key))
		}
	case bool:
		if _, ok := actual.(bool); !ok {
			errs = append(errs, fmt.Sprintf("value for key '%v' expected to be of type boolean", key))
		}
	case float64:
		if _, ok := actual.(float64); !ok {
			errs = append(errs, fmt.Sprintf("value for key '%v' expected to be of type number", key))
		}
	case []interface{}:
		if actualArray, ok := actual.([]interface{}); !ok {
			errs = append(errs, fmt.Sprintf("value for key '%v' expected to be of type array", key))
		} else {
			errs = append(errs, validateArray(key, expected, actualArray)...)
		}
	case map[string]interface{}:
		if actualObj, ok := actual.(map[string]interface{}); !ok {
			errs = append(errs, fmt.Sprintf("value for key '%v' expected to be of type object", key))
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
