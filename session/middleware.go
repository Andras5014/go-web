package session

import (
	"github.com/Andras5014/go-web"
	"net/http"
)

// NeedSession 返回一个中间件函数，用于检查请求中是否存在有效的会话
// 如果会话不存在，则调用 lossSessHandler 处理函数，通常用于处理会话丢失的情况
// 如果 lossSessHandler 为空，则默认返回状态码为 401 的未授权响应
// 如果存在有效的会话，则继续执行后续的处理函数，并在处理完后检查会话是否被修改
// 如果被修改则保存会话到存储中
func NeedSession(m *Manager, lossSessHandler web.HandleFunc) web.HandleFunc {
	// 如果 lossSessHandler 为空，则设置默认处理函数
	if lossSessHandler == nil {
		lossSessHandler = func(ctx *web.Context) {
			ctx.Status(http.StatusUnauthorized)
			ctx.Abort()
		}
	}
	// 返回中间件函数
	return func(ctx *web.Context) {
		sess, err := m.GetSession(ctx)
		if err != nil {
			// 如果会话不存在，则调用处理函数处理丢失会话的情况
			lossSessHandler(ctx)
			return
		}

		// 继续执行后续的处理函数
		ctx.Next()
		// 如果会话被修改，则保存会话到存储中
		if sess.Modified() {
			err = m.SaveSession(ctx, sess)
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				ctx.Abort()
			}
		}
	}
}
