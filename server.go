package web

import (
	"net"
	"net/http"
)

// HandleFunc 路由处理函数
type HandleFunc func(ctx *Context)

var _ Server = &Engine{}

// Server 服务器的基本功能
type Server interface {
	http.Handler
	Start(addr string) error
	Handle(method string, path string, handlers ...HandleFunc)
}

// EngineOption 引擎的可选项类型
type EngineOption func(*Engine)

// Engine 实现了 Server 接口
type Engine struct {
	*router                              //继承路由
	RouterGroup                          //包含默认路由组
	NotFoundHandler HandleFunc           // 404 处理函数
	AfterStart      func(l net.Listener) // 启动后回调
}

// DefaultNotFoundHandler 默认的404页面处理函数
var DefaultNotFoundHandler = func(ctx *Context) {
	ctx.StatusCode = http.StatusNotFound
	ctx.RespData = []byte("404 page not found")
}

// NewEngine 创建一个新的引擎实例
func NewEngine(opts ...EngineOption) *Engine {
	res := &Engine{
		router: newRouter(),
		RouterGroup: RouterGroup{
			basePath: "/",
		},
		NotFoundHandler: DefaultNotFoundHandler,
	}
	res.RouterGroup.engine = res
	for _, opt := range opts {
		opt(res)
	}
	return res
}

// WithNotFoundHandler 设置引擎的404处理函数
func WithNotFoundHandler(h HandleFunc) EngineOption {
	return func(e *Engine) {
		e.NotFoundHandler = h
	}
}

// WithAfterStart 设置引擎启动后的回调函数
func WithAfterStart(h func(l net.Listener)) EngineOption {
	return func(e *Engine) {
		e.AfterStart = h
	}
}

// ServeHTTP 实现了http.Handler接口的ServeHTTP方法
func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := newContext(writer, request)
	e.serve(ctx)
}

// serve 处理请求的核心方法
func (e *Engine) serve(ctx *Context) {
	// 查找路由，如果未找到则调用默认的404处理函数，否则执行对应的处理函数链
	info, ok := e.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info.node.handlers == nil {
		e.NotFoundHandler(ctx)
	} else {
		ctx.MatchedRoute = info.node.route
		ctx.PathParams = info.pathParams
		ctx.handlers = info.node.handlers
		ctx.Next()
	}
	// 发送HTTP响应
	e.flushResp(ctx)
}

// flushResp 发送HTTP响应
func (e *Engine) flushResp(ctx *Context) {
	ctx.Resp.WriteHeader(ctx.StatusCode)
	if ctx.RespData != nil {
		_, _ = ctx.Resp.Write(ctx.RespData)
	}
}

// Start 启动服务器
func (e *Engine) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 这里可以执行after start的操作
	if e.AfterStart != nil {
		e.AfterStart(l)
	}
	return http.Serve(l, e)
}

// Handle 注册路由处理函数
func (e *Engine) Handle(method string, path string, handlers ...HandleFunc) {
	if len(handlers) == 0 || handlers[0] == nil {
		panic("HandleFunc is empty")
	}
	e.router.addRoute(method, path, handlers...)
}
