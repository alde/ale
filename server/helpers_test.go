package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_writeJSON(t *testing.T) {
	wr := httptest.NewRecorder()

	writeJSON(200, "foo", wr)
	assert.Equal(t, contentTypeJSON, wr.HeaderMap["Content-Type"][0])
	assert.Equal(t, http.StatusOK, wr.Code)
}

func Test_notFound(t *testing.T) {
	wr := httptest.NewRecorder()

	notFound(wr)
	assert.Equal(t, contentTypeJSON, wr.HeaderMap["Content-Type"][0])
	assert.Equal(t, http.StatusNotFound, wr.Code)
}

func Test_writeError(t *testing.T) {
	wr := httptest.NewRecorder()

	writeError(http.StatusInternalServerError, "An error", wr)

	assert.Equal(t, contentTypeJSON, wr.HeaderMap["Content-Type"][0])
	assert.Equal(t, http.StatusInternalServerError, wr.Code)
}
