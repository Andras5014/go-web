package web

import (
	"encoding/json"
	"errors"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
)

// abortIndex 定义中止索引值
const abortIndex int = math.MaxInt8

// Context 上下文结构体，包含了请求和响应相关信息
type Context struct {
	Req          *http.Request       // HTTP请求
	Resp         http.ResponseWriter // HTTP响应
	PathParams   map[string]string   // 路径参数
	queryCache   url.Values          // 查询缓存
	MatchedRoute string              // 匹配到的路由
	Values       map[string]any

	index    int          // 处理函数索引
	handlers []HandleFunc // 处理函数列表

	StatusCode int    // 响应状态码
	RespData   []byte // 响应数据
}

// newContext 创建新的上下文实例
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Req:   req,
		Resp:  w,
		index: -1,
	}
}

func (c *Context) Get(key string) (any, bool) {
	if c.Values == nil {
		return nil, false
	}
	val, ok := c.Values[key]
	return val, ok
}

func (c *Context) Set(key string, val any) {
	if c.Values == nil {
		c.Values = make(map[string]any)
	}
	c.Values[key] = val
}

func (c *Context) Status(status int) {
	c.StatusCode = status
}

// JSON 发送JSON格式的响应
func (c *Context) JSON(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.Resp.Header().Set("Content-Type", "application/json")
	c.StatusCode = status
	c.RespData = data
	return nil
}

// String 发送文本响应
func (c *Context) String(status int, val string) error {
	c.Resp.Header().Set("Content-Type", "text/plain")
	c.StatusCode = status
	c.RespData = []byte(val)
	return nil
}

// HTML 发送HTML响应
func (c *Context) HTML(status int, val string) error {
	c.Resp.Header().Set("Content-Type", "text/html")
	c.StatusCode = status
	c.RespData = []byte(val)
	return nil
}

// JsonOK 发送HTTP OK状态的JSON响应
func (c *Context) JsonOK(val any) error {
	return c.JSON(http.StatusOK, val)
}

// SetCookie 设置Cookie
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp, cookie)
}

// GetCookie 获取Cookie
func (c *Context) GetCookie(name string) (*http.Cookie, bool) {
	cookie, err := c.Req.Cookie(name)
	if err != nil {
		return nil, false
	}
	return cookie, true
}

// BindJSON 解析JSON数据
func (c *Context) BindJSON(val any) error {
	if val == nil {
		return errors.New("nil pointer")
	}
	return json.NewDecoder(c.Req.Body).Decode(val)
}

// Param 获取路径参数
func (c *Context) Param(key string) string {
	return c.PathParams[key]
}

// FormValue 获取表单值
func (c *Context) FormValue(key string) (string, bool) {
	err := c.Req.ParseForm()
	if err != nil {
		return "", false
	}
	vals, ok := c.Req.Form[key]
	if !ok {
		return "", false
	}
	return vals[0], true
}

// MultipartForm 获取Multipart表单
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Req.ParseMultipartForm(32 << 20)
	if err != nil {
		return nil, err
	}
	return c.Req.MultipartForm, nil
}

// QueryValue 获取查询参数值
func (c *Context) QueryValue(key string) (string, bool) {
	if c.queryCache == nil {
		c.queryCache = c.Req.URL.Query()
	}
	vals, ok := c.queryCache[key]
	if !ok {
		return "", false
	}
	return vals[0], true
}

// PathValue 获取路径参数值
func (c *Context) PathValue(key string) (string, bool) {
	res, ok := c.PathParams[key]
	return res, ok
}

// Next 执行下一个处理函数
func (c *Context) Next() {
	c.index++
	for n := len(c.handlers); c.index < n; c.index++ {
		c.handlers[c.index](c)
	}
}

// Abort 中止请求处理
func (c *Context) Abort() {
	c.index = abortIndex
}

// IsAborted 判断请求是否已中止
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}
