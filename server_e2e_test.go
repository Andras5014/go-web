//go:build e2e

package web

import (
	"fmt"
	"testing"
)

func TestServer_e2e(t *testing.T) {
	e := NewEngine()

	e.GET("/user", func(ctx *Context) {
		fmt.Println("GET /user handler1")
	}, func(ctx *Context) {
		fmt.Println("GET /user handler2")
	})

	e.GET("/user/:id", func(ctx *Context) {
		id, ok := ctx.PathValue("id")
		if !ok {
			ctx.Resp.WriteHeader(400)
			return
		}
		_ = ctx.JSON(200, map[string]string{"id": id})
	})

	err := e.Start(":8080")
	if err != nil {
		t.Fatalf("server start failed: %v", err)
	}
}
