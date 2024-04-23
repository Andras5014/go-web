package web

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// MiddlewareBuilder 定义中间件构建器的接口
type MiddlewareBuilder interface {
	Build() HandleFunc
}

// DefaultLogFunc 默认的日志函数
func DefaultLogFunc(info string) {
	log.Println(info)
}

// LoggerBuilder 日志中间件构建器
type LoggerBuilder struct {
	LogFunc func(log string)
}

// Build 构建Logger中间件
func (l LoggerBuilder) Build() HandleFunc {
	if l.LogFunc == nil {
		l.LogFunc = DefaultLogFunc
	}
	return func(ctx *Context) {
		startTime := time.Now()
		ctx.Next()
		defer func() {
			al := accessLog{
				Host:    ctx.Req.Host,
				Route:   ctx.MatchedRoute,
				Method:  ctx.Req.Method,
				Path:    ctx.Req.URL.Path,
				Latency: time.Since(startTime),
			}
			data, _ := json.Marshal(al)
			l.LogFunc(string(data))
		}()
	}
}

// accessLog 访问日志结构体
type accessLog struct {
	Host    string        // 主机
	Route   string        // 路由
	Method  string        // HTTP方法
	Path    string        // 请求路径
	Latency time.Duration //响应时间
}

// PrometheusBuilder Prometheus监控中间件构建器
type PrometheusBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

// Build 构建Prometheus监控中间件
func (p PrometheusBuilder) Build() HandleFunc {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: p.Namespace,
		Subsystem: p.Subsystem,
		Name:      p.Name,
		Help:      p.Help,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"pattern", "method", "status"})

	prometheus.MustRegister(vector)

	return func(ctx *Context) {
		startTime := time.Now()
		defer func() {
			pattern := ctx.MatchedRoute
			if pattern == "" {
				pattern = "unknown"
			}
			vector.WithLabelValues(pattern, ctx.Req.Method, strconv.Itoa(ctx.StatusCode)).
				Observe(float64(time.Since(startTime).Milliseconds()))
		}()
		ctx.Next()
	}
}

// RecoverBuilder 恢复中间件构建器
type RecoverBuilder struct {
	LogFunc  func(log string) // 日志函数
	LogStack bool             // 是否记录堆栈信息
	Handler  HandleFunc       // 回复处理函数
}

// DefaultRecoverHandler 默认的恢复处理函数
func DefaultRecoverHandler(ctx *Context) {
	ctx.StatusCode = 500
	ctx.RespData = []byte("Internal Server Error")
}

// Build 构建Recover中间件
func (r RecoverBuilder) Build() HandleFunc {
	if r.LogFunc == nil {
		r.LogFunc = DefaultLogFunc
	}
	if r.Handler == nil {
		r.Handler = DefaultRecoverHandler
	}
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				if r.LogStack {
					r.LogFunc(trace(fmt.Sprintf("%s", err)))
				} else {
					r.LogFunc(fmt.Sprintf("%s", err))
				}
				r.Handler(ctx)
			}
		}()
		ctx.Next()
	}
}

// trace 追踪错误信息的堆栈
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
