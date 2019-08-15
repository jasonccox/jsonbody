[![GoDoc](https://godoc.org/gitlab.com/jasonccox/jsonbody?status.svg)](https://godoc.org/gitlab.com/jasonccox/jsonbody)

# jsonbody

jsonbody is a Golang middleware library that makes receiving JSON web request bodies, validating them, and sending JSON response bodies easy.

## Installation

Just run `go get gitlab.com/jasonccox/jsonbody` and start using [`jsonbody.Middleware`](https://godoc.org/gitlab.com/jasonccox/jsonbody#Middleware) in your routes!

## Examples

### Using the `jsonbody.Middleware` in a Route

The following code creates a `jsonbody.Middleware` that will ensure all POST requests to the `/turtle` route have a body matching the given schema. Requests of any other type can have any JSON body.

```go
func main() {

	// create the middleware
	middleware, err := jsonbody.NewMiddleware(handler{}, map[string]string{
		http.MethodPost: `{
			"name": "",         	// required key "name" with string value
			"age": 0,           	// required key "age" with number value
			"details": {        	// required key "details" with object value
				"?species": "",     // optional key "species" with string value
				"aquatic": false    // required key "aquatic" with boolean value
			},
			"children": ["name"]	// required key children with array of string values
		}`,
	})
	if err != nil {
		log.Fatalf("creating jsonbody.Middleware failed: %v\n", err)
	}

	// use the middleware in the route
	http.Handle("/turtle", middleware)

	...
}
```

### Handling the Request

The following code uses the `jsonbody.Reader` and `jsonbody.Writer` to handle a POST request.

```go
type handler struct {}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonWriter, ok := w.(jsonbody.Writer)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonReader, ok := r.Body.(jsonbody.Reader)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonBody := jsonReader.JSON() // returns a map[string]interface{} representing the request body

	turt := turtle{
		name:     jsonBody["name"].(string), // we can safely assert the type because the middleware already checked it
		age:      jsonBody["age"].(float64), // JSON numbers are represented as float64
		children: jsonBody["children"].([]string)
	}

	turtDetails := jsonBody["details"].(map[string]interface{}) // JSON objects are represented as map[string]interface{}
	turt.aquatic = turtDetails["aquatic"].(bool)

	if s, ok := turtDetails["species"]; ok { // details.species was optional, so we need to make sure it was set before using it
		turt.species = s.(string)
	}

	if turt.age < 0.0 {
		jsonWriter.WriteErrors(http.StatusBadRequest, "age must be positive") // sends back an error body
		return
	}

	// do some processing here...

	jsonWriter.WriteJSON(http.StatusCreated, turt) // converts turt to JSON and writes it as the response body
}
```

## Documentation

For more in-depth information, check out the [godoc](https://godoc.org/gitlab.com/jasonccox/jsonbody).