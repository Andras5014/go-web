package web

import (
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	var h Server
	http.ListenAndServe(":8080", h)
}
