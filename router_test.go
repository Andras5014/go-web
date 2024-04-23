package web

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

// mockHandler 模拟处理函数
var mockHandler HandleFunc = func(ctx *Context) {}

func TestRouter_addRouter(t *testing.T) {
	testRoutes := []struct {
		method   string
		path     string
		handlers []HandleFunc
	}{
		{
			method: http.MethodGet,
			path:   "/user/home",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodGet,
			path:   "/",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodGet,
			path:   "/user",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodPost,
			path:   "/",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodPost,
			path:   "/order",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodPost,
			path:   "/order/detail",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method:   http.MethodPost,
			path:     "/order/detail/:id",
			handlers: []HandleFunc{mockHandler},
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodPost,
			path:   "/order/detail1",
			handlers: []HandleFunc{
				mockHandler,
			},
		},

		{
			method: http.MethodPost,
			path:   "origin",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
	}

	r := newRouter()

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path:     "/",
				handlers: []HandleFunc{mockHandler},
				children: []*node{
					{
						path:     "user",
						handlers: []HandleFunc{mockHandler},
						children: []*node{
							{
								path:     "home",
								handlers: []HandleFunc{mockHandler},
							},
						},
					},
				},
			},

			http.MethodPost: {
				path:     "/",
				handlers: []HandleFunc{mockHandler},
				children: []*node{

					{
						path:     "order",
						handlers: []HandleFunc{mockHandler},
						startChild: &node{
							path:     "*",
							handlers: []HandleFunc{mockHandler},
						},
						children: []*node{
							{
								path:     "detail",
								handlers: []HandleFunc{mockHandler},
								paramChild: &node{
									path:     ":id",
									handlers: []HandleFunc{mockHandler},
								},
							},
							{
								path:     "detail1",
								handlers: []HandleFunc{mockHandler},
							},
						},
					},
					{
						path: "origin",
						handlers: []HandleFunc{
							mockHandler,
						},
					},
				},
			},
		},
	}
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, route.handlers...)
	}

	if msg, equal := r.equal(wantRouter); !equal {
		t.Errorf("router 不相同: %s", msg)
	}

	// 普通节点重复注册panic
	r = newRouter()
	r.addRoute(http.MethodGet, "/user/name", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/user/name", mockHandler)
	}, "duplicated path")

	// 根节点重复注册panic
	r = newRouter()
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	}, "duplicated path")

	// 通配符节点重复注册panic
	r = newRouter()
	r.addRoute(http.MethodGet, "/user/:id", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/user/*", mockHandler)
	}, "can't register wildcard and param node at the same time")

	// 非同名路径参数节点重复注册panic
	r = newRouter()
	r.addRoute(http.MethodGet, "/user/:id", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/user/:name", mockHandler)
	}, "can't register two param nodes at the same level")
}

func (n *node) equal(y *node) (string, bool) {

	if y == nil {
		return "y is nil", false
	}

	if n.path != y.path {
		return "path 不相同", false
	}
	if len(n.children) != len(y.children) {
		return "children 数量不相同", false
	}

	if len(n.handlers) != len(y.handlers) {
		return "handlers 数量不相同", false
	}

	if n.startChild != nil {
		msg, equal := n.startChild.equal(y.startChild)
		if !equal {
			return msg, false
		}
	}

	if n.paramChild != nil {
		msg, equal := n.paramChild.equal(y.paramChild)
		if !equal {
			return msg, false
		}
	}

	for i := 0; i < len(n.handlers); i++ {
		nHandler := reflect.ValueOf(n.handlers[i])
		yHandler := reflect.ValueOf(y.handlers[i])
		if nHandler != yHandler {
			return "handler 不相同", false
		}
	}

	for i := 0; i < len(n.children); i++ {
		msg, equal := n.children[i].equal(y.children[i])
		if !equal {
			return msg, false
		}
	}
	return "", true
}

func (r *router) equal(y *router) (string, bool) {
	for k, v := range r.trees {
		dst, ok := y.trees[k]
		if !ok {
			return "找不到对应的http method", false
		}
		msg, equal := v.equal(dst)
		if !equal {
			return msg, false
		}
	}

	return "", true
}

func TestRouter_findRoute(t *testing.T) {
	r := newRouter()

	testRoutes := []struct {
		method   string
		path     string
		handlers []HandleFunc
	}{
		{
			method: http.MethodGet,
			path:   "/user/home",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodGet,
			path:   "/",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodGet,
			path:   "/user",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodPost,
			path:   "/",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodPost,
			path:   "/order",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodPost,
			path:   "/order/detail",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method:   http.MethodPost,
			path:     "/order/detail/:id",
			handlers: []HandleFunc{mockHandler},
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
		{
			method: http.MethodPost,
			path:   "/order/detail1",
			handlers: []HandleFunc{
				mockHandler,
			},
		},

		{
			method: http.MethodPost,
			path:   "origin",
			handlers: []HandleFunc{
				mockHandler,
			},
		},

		{
			method: http.MethodPost,
			path:   "/post/*/detail",
			handlers: []HandleFunc{
				mockHandler,
			},
		},
	}

	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		info      *matchInfo
	}{
		{
			name:      "order detail",
			method:    http.MethodPost,
			path:      "/order/detail",
			wantFound: true,
			info: &matchInfo{
				node: &node{
					path:     "detail",
					handlers: []HandleFunc{mockHandler},
				},
				pathParams: map[string]string{},
			},
		},

		{
			name:      "order",
			method:    http.MethodPost,
			path:      "/order",
			wantFound: true,
			info: &matchInfo{
				node: &node{
					path:     "order",
					handlers: []HandleFunc{mockHandler},
					children: []*node{
						{
							path:     "detail",
							handlers: []HandleFunc{mockHandler},
						},
					},
				},
				pathParams: map[string]string{},
			},
		},
		{
			name:      "order *",
			method:    http.MethodPost,
			path:      "/order/123",
			wantFound: true,
			info: &matchInfo{
				node: &node{
					path:     "*",
					handlers: []HandleFunc{mockHandler},
				},
				pathParams: map[string]string{},
			},
		},
		{
			name:      "order detail :id",
			method:    http.MethodPost,
			path:      "/order/detail/123",
			wantFound: true,
			info: &matchInfo{
				node: &node{
					path:     ":id",
					handlers: []HandleFunc{mockHandler},
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			name:      "post * detail",
			method:    http.MethodPost,
			path:      "/post/123/detail",
			wantFound: true,
			info: &matchInfo{
				node: &node{
					path:     "detail",
					handlers: []HandleFunc{mockHandler},
				},
				pathParams: map[string]string{},
			},
		},

		{
			name:      "not found",
			method:    http.MethodGet,
			path:      "/as",
			wantFound: false,
		},

		{
			name:      "order/*/x not found",
			method:    http.MethodPost,
			path:      "/order/123/x",
			wantFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}
			assert.Equal(t, tc.info.node.path, n.node.path)
			assert.Equal(t, tc.info.pathParams, n.pathParams)
		})
	}
}
