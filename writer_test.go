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

	err := w.WriteJSON(200, "hi")
	assert.Equal(t, nil, err)

	err = w.WriteJSON(200, "hello")
	assert.NotEqual(t, nil, err)
}

func TestWriteJSONReturnsErrIfWriteFails(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Once().Return(0, errors.New("error"))
	mockRW.On("Header", mock.Anything).Return(http.Header{})
	mockRW.On("WriteHeader", mock.Anything).Return()

	err := w.WriteJSON(200, "hi")
	assert.NotEqual(t, nil, err)
}

func TestWriteJSONAllowsMultipleCallsIfErr(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Once().Return(0, errors.New("error"))
	mockRW.On("Write", mock.Anything).Once().Return(1, nil)

	mockRW.On("Header", mock.Anything).Return(http.Header{})
	mockRW.On("WriteHeader", mock.Anything).Return()

	err := w.WriteJSON(200, "hi")
	assert.NotEqual(t, nil, err)

	err = w.WriteJSON(200, "hello")
	assert.Equal(t, nil, err)
}

func TestWriteJSONWritesContentTypeHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := Writer{ResponseWriter: recorder}

	err := w.WriteJSON(200, "hello")
	assert.Equal(t, nil, err)

	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestWriteJSONWritesStatusCode(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := Writer{ResponseWriter: recorder}

	err := w.WriteJSON(200, "hello")
	assert.Equal(t, nil, err)

	assert.Equal(t, 200, recorder.Code)
}

func TestWriteJSONWritesJSON(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Return(1, nil)
	mockRW.On("Header", mock.Anything).Return(http.Header{})
	mockRW.On("WriteHeader", mock.Anything).Return()

	err := w.WriteJSON(200, map[string]string{"key": "value"})
	assert.Equal(t, nil, err)

	assert.Equal(t, []byte(`{"key":"value"}`), mockRW.lastBytes)
}

func TestWriteErrorsReturnsErrIfCalledTwice(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := Writer{ResponseWriter: recorder}

	err := w.WriteErrors(400, "hi")
	assert.Equal(t, nil, err)

	err = w.WriteErrors(400, "hello")
	assert.NotEqual(t, nil, err)
}

func TestWriteErrorsReturnsErrIfWriteFails(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Once().Return(0, errors.New("error"))

	mockRW.On("Header", mock.Anything).Return(http.Header{})
	mockRW.On("WriteHeader", mock.Anything).Return()

	err := w.WriteErrors(400, "hi")
	assert.NotEqual(t, nil, err)
}

func TestWriteErrorsAllowsMultipleCallsIfErr(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Once().Return(0, errors.New("error"))
	mockRW.On("Write", mock.Anything).Once().Return(1, nil)

	mockRW.On("Header", mock.Anything).Return(http.Header{})
	mockRW.On("WriteHeader", mock.Anything).Return()

	err := w.WriteErrors(400, "hi")
	assert.NotEqual(t, nil, err)

	err = w.WriteErrors(400, "hello")
	assert.Equal(t, nil, err)
}

func TestWriteErrorsWritesContentTypeHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := Writer{ResponseWriter: recorder}

	err := w.WriteErrors(400, "hello")
	assert.Equal(t, nil, err)

	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestWriteErrorsWritesStatusCode(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := Writer{ResponseWriter: recorder}

	err := w.WriteErrors(400, "hello")
	assert.Equal(t, nil, err)

	assert.Equal(t, 400, recorder.Code)
}

func TestWriteErrorsWritesOneError(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Return(1, nil)
	mockRW.On("Header", mock.Anything).Return(http.Header{})
	mockRW.On("WriteHeader", mock.Anything).Return()

	err := w.WriteErrors(400, "error")
	assert.Equal(t, nil, err)

	assert.Equal(t, []byte(`{"errors":["error"]}`), mockRW.lastBytes)
}

func TestWriteErrorsWritesMultipleErrors(t *testing.T) {
	mockRW := mockResponseWriter{}
	w := Writer{ResponseWriter: &mockRW}

	mockRW.On("Write", mock.Anything).Return(1, nil)
	mockRW.On("Header", mock.Anything).Return(http.Header{})
	mockRW.On("WriteHeader", mock.Anything).Return()

	err := w.WriteErrors(400, "error1", "error2", "error3")
	assert.Equal(t, nil, err)

	assert.Equal(t, []byte(`{"errors":["error1","error2","error3"]}`), mockRW.lastBytes)
}
