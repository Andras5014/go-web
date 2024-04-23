package web

import (
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

var _ Server = &Engine{}

type Server interface {
	http.Handler
	Start(addr string) error
	Handle(method string, path string, handlers ...HandleFunc)
}

type EngineOption func(*Engine)

type Engine struct {
	*router
	RouterGroup
	NotFoundHandler HandleFunc
	AfterStart      func(l net.Listener)
}

var DefaultNotFoundHandler = func(ctx *Context) {
	ctx.StatusCode = http.StatusNotFound
	ctx.RespData = []byte("404 page not found")
}

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

func WithNotFoundHandler(h HandleFunc) EngineOption {
	return func(e *Engine) {
		e.NotFoundHandler = h
	}
}

func WithAfterStart(h func(l net.Listener)) EngineOption {
	return func(e *Engine) {
		e.AfterStart = h
	}
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := newContext(writer, request)
	e.serve(ctx)
}

func (e *Engine) serve(ctx *Context) {
	info, ok := e.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info.node.handlers == nil {
		e.NotFoundHandler(ctx)
	} else {
		ctx.MatchedRoute = info.node.route
		ctx.PathParams = info.pathParams
		ctx.handlers = info.node.handlers
		ctx.Next()
	}

	e.flushResp(ctx)
}

func (e *Engine) flushResp(ctx *Context) {
	ctx.Resp.WriteHeader(ctx.StatusCode)
	if ctx.RespData != nil {
		_, _ = ctx.Resp.Write(ctx.RespData)
	}
}

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

func (e *Engine) Handle(method string, path string, handlers ...HandleFunc) {
	if len(handlers) == 0 || handlers[0] == nil {
		panic("HandleFunc is empty")
	}
	e.router.addRoute(method, path, handlers...)
}
