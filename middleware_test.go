package web

import (
	"net/http"
	"testing"
	"time"
)

func TestLoggerBuilder_Build(t *testing.T) {
	server := NewEngine()
	builder := LoggerBuilder{
		LogFunc: func(log string) {
			t.Log(log)
		},
	}
	server.Use(builder.Build())

	server.GET("/test", func(c *Context) {
		time.Sleep(1 * time.Second)
		_ = c.String(200, "test")
	})

	mockRequest, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	server.ServeHTTP(&MockWriter{}, mockRequest)

}

func TestRecoverBuilder_Build(t *testing.T) {
	server := NewEngine()
	builder := RecoverBuilder{
		LogStack: true,
	}
	server.Use(builder.Build())

	server.GET("/test", func(c *Context) {
		panic("test")
	})

	mockRequest, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	server.ServeHTTP(&MockWriter{}, mockRequest)
}
