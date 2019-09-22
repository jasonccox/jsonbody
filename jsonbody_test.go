package jsonbody

import (
	"errors"
	"io/ioutil"
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

func (m *mockReader) Read(p []byte) (int, error) {
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

func TestServeHTTPIgnoresWrongContentTypeIfNoSchemaSet(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	mw := &middleware{next: next}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Content-Type", "text/html")
	mw.ServeHTTP(recorder, request)

	assert.Equal(t, 200, recorder.Code)
}

func TestServeHTTPSends400IfWrongContentTypeAndSchemaSet(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	mw := &middleware{
		next:   next,
		schema: make(map[string]interface{}),
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Content-Type", "text/html")
	mw.ServeHTTP(recorder, request)

	assert.Equal(t, 400, recorder.Code)
}

func TestServeHTTPSendsErrorsIfWrongContentTypeAndSchemaSet(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	mw := &middleware{
		next:   next,
		schema: make(map[string]interface{}),
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Content-Type", "text/html")
	mw.ServeHTTP(recorder, request)

	body := make([]byte, recorder.Body.Len())
	recorder.Body.Read(body)

	assert.Equal(t, `{"errors":["content type must be application/json"]}`, string(body))
}

func TestServeHTTPNotCallNextIfWrongContentTypeAndSchemaSet(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	mw := &middleware{
		next:   next,
		schema: make(map[string]interface{}),
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Content-Type", "text/html")
	mw.ServeHTTP(recorder, request)

	next.AssertNotCalled(t, "ServeHTTP", mock.Anything, mock.Anything)
}

func TestServeHTTPIgnoresEmptyBodyIfNoSchemaSet(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	mw := &middleware{next: next}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Content-Type", "application/json")
	mw.ServeHTTP(recorder, request)

	assert.Equal(t, 200, recorder.Code)
}

func TestServeHTTPSends400IfBodyEmptyAndSchemaSet(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	mw := &middleware{
		next:   next,
		schema: make(map[string]interface{}),
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Content-Type", "application/json")
	mw.ServeHTTP(recorder, request)

	assert.Equal(t, 400, recorder.Code)
}

func TestServeHTTPSendsErrorsIfBodyEmptyAndSchemaSet(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	mw := &middleware{
		next:   next,
		schema: make(map[string]interface{}),
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Content-Type", "application/json")
	mw.ServeHTTP(recorder, request)

	body := make([]byte, recorder.Body.Len())
	recorder.Body.Read(body)

	assert.Equal(t, `{"errors":["expected a JSON body"]}`, string(body))
}

func TestServeHTTPNotCallNextIfBodyEmptyAndSchemaSet(t *testing.T) {
	next := &mockHandler{}
	mw := &middleware{
		next:   next,
		schema: make(map[string]interface{}),
	}

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Content-Type", "application/json")
	mw.ServeHTTP(recorder, request)

	next.AssertNotCalled(t, "ServeHTTP", mock.Anything, mock.Anything)
}

func TestServeHTTPSends400IfBodyNotJSON(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	mw := &middleware{next: next}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	request.Header.Set("Content-Type", "application/json")
	mw.ServeHTTP(recorder, request)

	assert.Equal(t, 400, recorder.Code)
}

func TestServeHTTPSendsErrBodyIfBodyNotJSON(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	mw := &middleware{next: next}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	request.Header.Set("Content-Type", "application/json")
	mw.ServeHTTP(recorder, request)

	body := make([]byte, recorder.Body.Len())
	recorder.Body.Read(body)

	assert.Equal(t, `{"errors":["expected a JSON body"]}`, string(body))
}

func TestServeHTTPNotCallNextIfBodyNotJSON(t *testing.T) {
	next := &mockHandler{}
	mw := middleware{next: next}

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	request.Header.Set("Content-Type", "application/json")
	mw.ServeHTTP(recorder, request)

	next.AssertNotCalled(t, "ServeHTTP", mock.Anything, mock.Anything)
}

func TestServeHTTPSends500OnOtherError(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	mw := &middleware{next: next}

	reader := mockReader{}
	reader.On("Read", mock.Anything).Return(10, errors.New("some err"))

	req := httptest.NewRequest(http.MethodPost, "/", &reader)
	req.ContentLength = 1

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, req)

	assert.Equal(t, 500, recorder.Code)
}

func TestServeHTTPNotCallNextOnOtherError(t *testing.T) {
	next := &mockHandler{}
	mw := middleware{next: next}

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	reader := mockReader{}
	reader.On("Read", mock.Anything).Return(10, errors.New("some err"))

	req := httptest.NewRequest(http.MethodPost, "/", &reader)
	req.ContentLength = 1

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, req)

	next.AssertNotCalled(t, "ServeHTTP", mock.Anything, mock.Anything)
}

func TestServeHTTPCallsNextCorrectly(t *testing.T) {
	next := &mockHandler{}
	mw := middleware{next: next}

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))

	next.AssertCalled(t, "ServeHTTP", mock.AnythingOfType("Writer"), mock.AnythingOfType("*http.Request"))

	reader, ok := next.Calls[0].Arguments.Get(1).(*http.Request).Body.(Reader)
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{}, reader.JSON())
}

func TestServeHTTPSends400IfBodyNotMatchSchema(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	schema, _ := parseSchema(`{ "s": "" }`)
	mw := middleware{
		next:   next,
		schema: schema,
	}

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))

	assert.Equal(t, 400, recorder.Code)
}

func TestServeHTTPSendsErrorsIfBodyNotMatchSchema(t *testing.T) {
	next := &mockHandler{}
	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()
	schema, _ := parseSchema(`{ "s": "" }`)
	mw := middleware{
		next:   next,
		schema: schema,
	}

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))

	assert.NotEqual(t, 0, recorder.Body.Len())
}

func TestServeHTTPNotCallNextIfBodyNotMatchSchema(t *testing.T) {
	next := &mockHandler{}
	schema, _ := parseSchema(`{ "s": "" }`)
	mw := middleware{
		next:   next,
		schema: schema,
	}

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))

	next.AssertNotCalled(t, "ServeHTTP", mock.Anything, mock.Anything)
}

func TestServeHTTPResetsBody(t *testing.T) {
	next := &mockHandler{}
	mw := middleware{next: next}

	next.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))

	receivedBody, err := ioutil.ReadAll(next.Calls[0].Arguments.Get(1).(*http.Request).Body)
	assert.Nil(t, err)
	assert.Equal(t, "{}", string(receivedBody))
}

func TestNewMiddlewareAddsParsedSchemaToHandler(t *testing.T) {
	mw := NewMiddleware(`{"schema": "s"}`)
	next := &mockHandler{}
	handler := mw(next).(*middleware)

	expectedSchema, _ := parseSchema(`{"schema": "s"}`)
	assert.Equal(t, expectedSchema, handler.schema)
}

func TestNewMiddlewareAddsNextToHandler(t *testing.T) {
	mw := NewMiddleware("")
	next := &mockHandler{}
	handler := mw(next).(*middleware)

	assert.Equal(t, next, handler.next)
}

func TestNewMiddlewarePanicsIfInvalidSchema(t *testing.T) {
	shouldPanic := func() {
		NewMiddleware("not json")
	}

	assert.Panics(t, shouldPanic)
}
