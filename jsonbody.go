package jsonbody

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Middleware converts the request body to a map and allows the response to be
// written as JSON. The request body can be fetched from the request context
// using ReqKey as the key and will be of type *map[string]interface{}. The
// response body can be send by calling Write().
type Middleware struct {
	Next      http.Handler
	reqSchema map[string]interface{}
}

var (
	errServerErr = errors.New("an unexpected error occurred")
	errBadBody   = errors.New("the body of the request was bad")
)

// NewMiddleware creates a new instance of a jsonbody Middleware. See the documentation
// for the method SetRequestSchema regarding schemaJSON.
func NewMiddleware(next http.Handler, schemaJSON []byte) (*Middleware, error) {
	m := Middleware{Next: next}
	err := m.SetRequestSchema(schemaJSON)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.Next == nil {
		m.Next = http.DefaultServeMux
	}

	writer := Writer{ResponseWriter: w}

	jsonBody, err := decodeBody(r)
	switch {
	case err == errBadBody:
		writer.WriteHeader(400)
		writer.WriteErrors("expected a JSON body")
		return
	case err == errServerErr:
		fallthrough
	case err != nil:
		log.Println(fmt.Errorf("jsonbody: failed to decode body: %v", err))
		writer.WriteHeader(500)
		return
	}

	reader := Reader{
		ReadCloser: r.Body,
		json:       jsonBody,
	}
	r.Body = reader

	m.Next.ServeHTTP(writer, r)
}

func decodeBody(r *http.Request) (map[string]interface{}, error) {
	if r.ContentLength == 0 {
		return nil, errBadBody
	}

	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil && err != io.EOF {
		log.Println(fmt.Errorf("jsonbody: failed to read entire body: %v", err))
		return nil, errServerErr
	}

	err = r.Body.Close()
	if err != nil {
		log.Println(fmt.Errorf("jsonbody: failed to close body: %v", err))
		return nil, errServerErr
	}

	var bodyJSON interface{}
	err = json.Unmarshal(body, &bodyJSON)
	if err != nil {
		log.Println(fmt.Errorf("jsonbody: failed to decode body: %v", err))
		return nil, errBadBody
	}

	return bodyJSON.(map[string]interface{}), nil
}
