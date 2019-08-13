package jsonbody

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

// Writer is an extension of a generic http.ResponseWriter. It provides the method
// WriteJSON for writing an object the the response body as JSON, as well as the
// method WriteErrors for easily writing errors to the response body.
type Writer struct {
	http.ResponseWriter
	written bool
}

// WriteJSON encodes an object as JSON and sends it as the response body, along
// with the Content-Type header. This method can only be called once, unless it
// returns an error.
func (w *Writer) WriteJSON(body interface{}) error {
	if w.written {
		return errors.New("method has already been called once and cannot be called again")
	}

	bytes, err := json.Marshal(body)
	if err != nil {
		log.Println(fmt.Errorf("jsonbody: failed to encode body: %v", err))
		return errors.New("encoding the response body as JSON failed")
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(bytes)
	if err != nil {
		log.Println(fmt.Errorf("jsonbody: failed to write body: %v", err))
		return errors.New("sending the response body failed")
	}

	w.written = true

	return nil
}

// WriteErrors encodes the given errors as a JSON array assigned to the key "errors"
// and sends it as the response body.
func (w *Writer) WriteErrors(errs ...string) error {
	err := w.WriteJSON(map[string][]string{
		"errors": errs,
	})

	return err
}
