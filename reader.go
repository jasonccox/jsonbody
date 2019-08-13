package jsonbody

import "io"

// Reader is an extension of a generic io.Reader. It provides the method JSON for
// retrieving the JSON request body as a map[string]interface{}.
type Reader struct {
	io.ReadCloser
	json map[string]interface{}
}

// JSON returns a a map[string]interface{} representing the request body. See the
// documentation for encoding/json regarding how the map represents the JSON data.
func (r Reader) JSON() map[string]interface{} {
	return r.json
}
