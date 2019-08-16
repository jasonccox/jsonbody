package jsonbody

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockReader struct {
	mock.Mock
}

func (m mockReader) Read(p []byte) (int, error) {
	returnVals := m.Called(p)
	return returnVals.Int(0), returnVals.Error(1)
}

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	m.Called(w, r)
}

func TestServeHTTPDefaultsToDefaultServeMux(t *testing.T) {
	mw, _ := NewMiddleware(nil, nil)

	mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))

	assert.Equal(t, http.DefaultServeMux, mw.Next)
}

func TestServeHTTPSends400IfBodyNotJSON(t *testing.T) {
	mw, _ := NewMiddleware(nil, nil)

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json")))

	assert.Equal(t, 400, recorder.Code)
}

func TestServeHTTPSendsErrBodyIfBodyNotJSON(t *testing.T) {
	mw, _ := NewMiddleware(nil, nil)

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json")))

	body := make([]byte, recorder.Body.Len())
	recorder.Body.Read(body)

	assert.Equal(t, `{"errors":["expected a JSON body"]}`, string(body))
}

func TestServeHTTPNotCallNextIfBodyNotJSON(t *testing.T) {
	next := &mockHandler{}
	mw := Middleware{Next: next}

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json")))

	next.AssertNotCalled(t, "ServeHTTP", mock.Anything, mock.Anything)
}

func TestServeHTTPSends500OnOtherError(t *testing.T) {
	mw, _ := NewMiddleware(nil, nil)

	reader := mockReader{}
	reader.On("Read", mock.Anything).Return(10, errors.New("some err"))

	req := httptest.NewRequest(http.MethodPost, "/", reader)
	req.ContentLength = 1

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, req)

	assert.Equal(t, 500, recorder.Code)
}

func TestServeHTTPNotCallNextOnOtherError(t *testing.T) {
	next := &mockHandler{}
	mw := Middleware{Next: next}

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	reader := mockReader{}
	reader.On("Read", mock.Anything).Return(10, errors.New("some err"))

	req := httptest.NewRequest(http.MethodPost, "/", reader)
	req.ContentLength = 1

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, req)

	next.AssertNotCalled(t, "ServeHTTP", mock.Anything, mock.Anything)
}

func TestServeHTTPCallsNextCorrectly(t *testing.T) {
	next := &mockHandler{}
	mw := Middleware{Next: next}

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))

	next.AssertCalled(t, "ServeHTTP", mock.AnythingOfType("Writer"), mock.AnythingOfType("*http.Request"))

	reader, ok := next.Calls[0].Arguments.Get(1).(*http.Request).Body.(Reader)
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{}, reader.JSON())
}

func TestServeHTTPSends400IfBodyNotMatchSchema(t *testing.T) {
	mw := Middleware{}
	mw.SetRequestSchema(http.MethodPost, `{ "s": "" }`)

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))

	assert.Equal(t, 400, recorder.Code)
}

func TestServeHTTPSendsErrBOdyIfBodyNotMatchSchema(t *testing.T) {
	mw := Middleware{}
	mw.SetRequestSchema(http.MethodPost, `{ "s": "" }`)

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))

	assert.NotEqual(t, 0, recorder.Body.Len())
}

func TestServeHTTPNotCallNextIfBodyNotMatchSchema(t *testing.T) {
	next := &mockHandler{}
	mw := Middleware{Next: next}
	mw.SetRequestSchema(http.MethodPost, `{ "s": "" }`)

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))

	next.AssertNotCalled(t, "ServeHTTP", mock.Anything, mock.Anything)
}

func TestServeHTTPValidateWithSchemaForMethod(t *testing.T) {
	next := &mockHandler{}
	mw := Middleware{Next: next}
	mw.SetRequestSchema(http.MethodGet, `{ "s": "" }`)

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))

	assert.Equal(t, 200, recorder.Code)
	next.AssertCalled(t, "ServeHTTP", mock.AnythingOfType("Writer"), mock.AnythingOfType("*http.Request"))
}

func TestNewMiddlewareSetsNext(t *testing.T) {
	next := http.FileServer(nil)

	mw, _ := NewMiddleware(next, nil)

	assert.Equal(t, next, mw.Next)
}

func TestNewMiddlewareSetsBodySchemas(t *testing.T) {
	bodySchemas := map[string]string{
		http.MethodGet:  `{ "a": false }`,
		http.MethodPost: `{ "b": 0 }`,
		http.MethodPut:  `{ "c": "s" }`,
	}

	m, err := NewMiddleware(nil, bodySchemas)
	assert.Nil(t, err)

	assert.Equal(t, map[string]interface{}{"a": false}, m.reqSchemas[http.MethodGet])
	assert.Equal(t, map[string]interface{}{"b": float64(0)}, m.reqSchemas[http.MethodPost])
	assert.Equal(t, map[string]interface{}{"c": "s"}, m.reqSchemas[http.MethodPut])
}

func TestNewMiddlewareReturnsErrIfAnyBodySchemasInvalid(t *testing.T) {
	bodySchemas := map[string]string{
		http.MethodGet:  `{ "a": false }`,
		http.MethodPost: `{ "b": 0`,
		http.MethodPut:  `{ "c": "s" }`,
	}

	m, err := NewMiddleware(nil, bodySchemas)
	assert.NotNil(t, err)
	assert.Equal(t, (*Middleware)(nil), m)
}

func TestNewMiddlewareNotSetSchemasIfNil(t *testing.T) {
	m, err := NewMiddleware(nil, nil)
	assert.Nil(t, err)

	assert.Equal(t, map[string]map[string]interface{}(nil), m.reqSchemas)
}
