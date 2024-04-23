//go:build e2e

package web

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestPrometheusBuilder_Build(t *testing.T) {

	b := &PrometheusBuilder{
		Namespace: "go_web",
		Subsystem: "web",
		Name:      "test",
		Help:      "test",
	}

	server := NewEngine()
	server.Use(b.Build())

	server.GET("/test", func(c *Context) {
		val := rand.Intn(1000) + 1
		time.Sleep(time.Duration(val) * time.Millisecond)
		_ = c.String(200, "test")
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			t.Log(err)
		}
	}()

	err := server.Start(":8080")
	t.Log(err)
}
