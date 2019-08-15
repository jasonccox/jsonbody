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
// written as JSON. When Middleware calls the Next.ServeHTTP(), it passes it a
// Writer and a *http.Request with Body set as a Reader. See documentation for
// Reader and Writer regarding accessing the request body and writing to the
// response body.
//
// Middleware can also optionally validate the request body by checking that its
// structure matches a pre-defined schema. If the request body does not match the
// schema, a 404 response with the following JSON body will be sent:
// 	{
//		"errors": [ <list of error strings> ]
//	}
type Middleware struct {
	Next       http.Handler
	reqSchemas map[string]map[string]interface{}
}

var (
	errServerErr = errors.New("an unexpected error occurred")
	errBadBody   = errors.New("the body of the request was bad")
)

// NewMiddleware creates a new instance of a Middleware.
//
// If next is nil, it will default to http.DefaultServeMux.
//
// bodySchemas maps HTTP request methods to the expected JSON body to be received.
// See the documentation for SetRequestSchema for more details.
func NewMiddleware(next http.Handler, bodySchemas map[string]string) (*Middleware, error) {
	m := &Middleware{Next: next}

	for method, schema := range bodySchemas {
		err := m.SetRequestSchema(method, schema)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.Next == nil {
		m.Next = http.DefaultServeMux
	}

	writer := Writer{ResponseWriter: w}

	jsonBody, err := decodeBody(r)
	switch {
	case err == errBadBody:
		writer.WriteHeader(http.StatusBadRequest)
		writer.WriteErrors("expected a JSON body")
		return
	case err == errServerErr:
		fallthrough
	case err != nil:
		log.Println(fmt.Errorf("jsonbody: failed to decode body: %v", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	errs := validateReqBody(m.reqSchemas[r.Method], jsonBody)
	if len(errs) > 0 {
		writer.WriteHeader(http.StatusBadRequest)
		writer.WriteErrors(errs...)
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

	// TODO: find a way to reset the body after reading

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
