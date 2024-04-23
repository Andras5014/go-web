package session

import "net/http"

// Propagator 定义了会话 ID 传播器的接口，用于在 HTTP 请求和响应之间传递会话 ID
type Propagator interface {
	Inject(id string, writer http.ResponseWriter) error // Inject 将会话 ID 注入到 HTTP 响应中
	Extract(req *http.Request) (string, error)          // Extract 从 HTTP 请求中提取会话 ID
	Clean(writer http.ResponseWriter) error             // Clean 清除已经注入的会话 ID
}

type CookiePropagatorOption func(propagator *CookiePropagator)

// CookiePropagator 实现了 Propagator 接口，用于通过 HTTP Cookie 传播会话 ID
type CookiePropagator struct {
	cookieName   string
	cookieOption func(cookie *http.Cookie)
}

func NewCookiePropagator() Propagator {
	return &CookiePropagator{
		cookieName: "session_id",
		cookieOption: func(cookie *http.Cookie) {

		},
	}
}

// WithCookieName 设置 CookiePropagator 的 Cookie 名称
func WithCookieName(name string) CookiePropagatorOption {
	return func(propagator *CookiePropagator) {
		propagator.cookieName = name
	}
}

// WithCookieOption 设置 CookiePropagator 的 Cookie 选项
func WithCookieOption(option func(cookie *http.Cookie)) CookiePropagatorOption {
	return func(propagator *CookiePropagator) {
		propagator.cookieOption = option
	}
}

// Inject 将会话 ID 注入到 HTTP 响应的 Cookie 中
func (c *CookiePropagator) Inject(id string, writer http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:  c.cookieName,
		Value: id,
		Path:  "/",
	}
	c.cookieOption(cookie)
	http.SetCookie(writer, cookie)
	return nil
}

// Extract 从 HTTP 请求中提取会话 ID
func (c *CookiePropagator) Extract(req *http.Request) (string, error) {
	cookie, err := req.Cookie(c.cookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// Clean 清除已经注入的会话 ID
func (c *CookiePropagator) Clean(writer http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:   c.cookieName,
		MaxAge: -1, // 立即删除 Cookie
	}
	http.SetCookie(writer, cookie)
	return nil
}
