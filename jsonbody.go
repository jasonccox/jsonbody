// Package jsonbody is a Golang middleware library that makes receiving JSON web
// request bodies, validating them, and sending JSON response bodies easy.
package jsonbody

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// NewMiddleware creates a middleware that converts the request body to a map and
// allows the response to be written as JSON. When Middleware calls the
// next.ServeHTTP(), it passes it a Writer and a *http.Request with Body set as a
// Reader. See documentation for Reader and Writer regarding accessing the request
// body and writing to the response body.
//
// The middleware can also optionally validate the content type and request body
// by checking that its structure matches a pre-defined schema. If the request
// body does not match the schema, a 400 response with the following JSON body
// will be sent:
// 	{
//		"errors": [ <list of error strings> ]
//	}
//
// The schemaJSON should essentially be a sample request body. All keys in the
// schemaJSON (unless they begin with a question mark) will be expected to be
// present in request bodies that pass through the Middleware. Additionally, all
// values will be expected to have the same type as the values in the schema.
// Arrays in the schema need only have one element in them against which all
// array elements in the real request will be verified. Finally, an empty object
// or empty array in the schema indicates that the object/array in the requests
// must be present but can have any contents. See the example below for further
// clarification.
//
// Setting schemaJSON to "" (the empty string) indicates that any JSON body
// (including none at all) and any content type should be accepted.
//
// Example Schema (don't actually include comments in yours)
// 	`{
//		"title": "", // body must contain a key "title" with a string value
//		"upvotes": 0, // body must contain a key "upvotes" with a number value
// 		"?public": false, // body may contain a key "public" with a boolean value
//		"comments": [ // body must contain a key "comments" with an array value
//			"" // each element in the "comments" array must be a string
//		],
//		"author": { // body must contain a key "author" with an object value
//			"name": "",	// "author" object must contain a key "name" with a string value
//			...
//		},
//		"metadata": {}, // body must contain a key "metadata" with an object value,
//		                // but the value can contain any keys, or none at all
//		"tags": [], // body must contain a key "tags" with an array value, but the
// 		            // elements can be of any type
//		...
//	}`
func NewMiddleware(schemaJSON string) func(next http.Handler) http.Handler {
	schemaMap, err := parseSchema(schemaJSON)
	if err != nil {
		panic("jsonbody: unexpected error while parsing schemaJSON: " + err.Error())
	}

	return func(next http.Handler) http.Handler {
		return &middleware{
			next:   next,
			schema: schemaMap,
		}
	}
}

var (
	errServerErr = errors.New("an unexpected error occurred")
	errBadBody   = errors.New("the body of the request was bad")
)

type middleware struct {
	next   http.Handler
	schema map[string]interface{}
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writer := Writer{ResponseWriter: w}

	if m.schema != nil && r.Header.Get("Content-Type") != "application/json" {
		writer.WriteErrors(http.StatusBadRequest, "content type must be application/json")
		return
	}

	body, err := decodeBody(r)
	switch {
	case err == errBadBody:
		writer.WriteErrors(http.StatusBadRequest, "expected a JSON body")
		return
	case err == errServerErr:
		fallthrough
	case err != nil:
		log.Println(fmt.Errorf("jsonbody: failed to decode body: %v", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	errs := validateReqBody(m.schema, body)
	if len(errs) > 0 {
		writer.WriteErrors(http.StatusBadRequest, errs...)
		return
	}

	reader := Reader{
		ReadCloser: r.Body,
		json:       body,
	}
	r.Body = reader

	m.next.ServeHTTP(writer, r)
}

func decodeBody(r *http.Request) (map[string]interface{}, error) {
	if r.ContentLength == 0 {
		return nil, nil // validateReqBody will determine whether an empty body is an error or not
	}

	body := make([]byte, r.ContentLength)
	defer r.Body.Close()
	_, err := r.Body.Read(body)
	if err != nil && err != io.EOF {
		log.Println(fmt.Errorf("jsonbody: failed to read entire body: %v", err))
		return nil, errServerErr
	}

	// reset body in case future handlers want to read it
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	var bodyJSON interface{}
	err = json.Unmarshal(body, &bodyJSON)
	if err != nil {
		log.Println(fmt.Errorf("jsonbody: failed to decode body: %v", err))
		return nil, errBadBody
	}

	return bodyJSON.(map[string]interface{}), nil
}
