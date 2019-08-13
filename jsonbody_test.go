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

type mockReader struct {
	mock.Mock
}

func (m mockReader) Read(p []byte) (int, error) {
	returnVals := m.Called(p)
	return returnVals.Int(0), returnVals.Error(1)
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

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
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
	assert.NotEqual(t, nil, reader.JSON())
}
