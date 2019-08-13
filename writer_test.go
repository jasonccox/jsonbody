package jsonbody

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockResponseWriter struct {
	mock.Mock
	lastBytes []byte
}

func (m *mockResponseWriter) Header() http.Header {
	returnVals := m.Called()
	return returnVals.Get(0).(http.Header)
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	m.lastBytes = b
	returnVals := m.Called(b)
	return returnVals.Int(0), returnVals.Error(1)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.Called(statusCode)
}

func TestWriteJSONReturnsErrIfCalledTwice(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := Writer{ResponseWriter: recorder}

	err := w.WriteJSON("hi")
	assert.Equal(t, nil, err)

	err = w.WriteJSON("hello")
	assert.NotEqual(t, nil, err)
}

func TestWriteJSONReturnsErrIfWriteFails(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Once().Return(0, errors.New("error!"))

	mockRW.On("Header", mock.Anything).Return(http.Header{})

	err := w.WriteJSON("hi")
	assert.NotEqual(t, nil, err)
}

func TestWriteJSONAllowsMultipleCallsIfErr(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Once().Return(0, errors.New("error!"))
	mockRW.On("Write", mock.Anything).Once().Return(1, nil)

	mockRW.On("Header", mock.Anything).Return(http.Header{})

	err := w.WriteJSON("hi")
	assert.NotEqual(t, nil, err)

	err = w.WriteJSON("hello")
	assert.Equal(t, nil, err)
}

func TestWriteJSONWritesContentTypeHeader(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Return(1, nil)

	header := http.Header{}

	mockRW.On("Header", mock.Anything).Return(header)

	err := w.WriteJSON("hello")
	assert.Equal(t, nil, err)

	assert.Equal(t, "application/json", header.Get("Content-Type"))
}

func TestWriteJSONWritesJSON(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Return(1, nil)
	mockRW.On("Header", mock.Anything).Return(http.Header{})

	err := w.WriteJSON(map[string]string{"key": "value"})
	assert.Equal(t, nil, err)

	assert.Equal(t, []byte("{\"key\":\"value\"}"), mockRW.lastBytes)
}

func TestWriteErrorsReturnsErrIfCalledTwice(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := Writer{ResponseWriter: recorder}

	err := w.WriteErrors("hi")
	assert.Equal(t, nil, err)

	err = w.WriteErrors("hello")
	assert.NotEqual(t, nil, err)
}

func TestWriteErrorsReturnsErrIfWriteFails(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Once().Return(0, errors.New("error!"))

	mockRW.On("Header", mock.Anything).Return(http.Header{})

	err := w.WriteErrors("hi")
	assert.NotEqual(t, nil, err)
}

func TestWriteErrorsAllowsMultipleCallsIfErr(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Once().Return(0, errors.New("error!"))
	mockRW.On("Write", mock.Anything).Once().Return(1, nil)

	mockRW.On("Header", mock.Anything).Return(http.Header{})

	err := w.WriteErrors("hi")
	assert.NotEqual(t, nil, err)

	err = w.WriteErrors("hello")
	assert.Equal(t, nil, err)
}

func TestWriteErrorsWritesContentTypeHeader(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Return(1, nil)

	header := http.Header{}

	mockRW.On("Header", mock.Anything).Return(header)

	err := w.WriteErrors("hello")
	assert.Equal(t, nil, err)

	assert.Equal(t, "application/json", header.Get("Content-Type"))
}

func TestWriteErrorsWritesOneError(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Return(1, nil)
	mockRW.On("Header", mock.Anything).Return(http.Header{})

	err := w.WriteErrors("error")
	assert.Equal(t, nil, err)

	assert.Equal(t, []byte("{\"errors\":[\"error\"]}"), mockRW.lastBytes)
}

func TestWriteErrorsWritesMultipleErrors(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Return(1, nil)
	mockRW.On("Header", mock.Anything).Return(http.Header{})

	err := w.WriteErrors("error1", "error2", "error3")
	assert.Equal(t, nil, err)

	assert.Equal(t, []byte("{\"errors\":[\"error1\",\"error2\",\"error3\"]}"), mockRW.lastBytes)
}
