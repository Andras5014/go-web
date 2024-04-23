package session

import (
	"github.com/Andras5014/go-web"
)

const sessionKey = "session"

// Manager 管理会话的结构体
type Manager struct {
	Propagator    Propagator // 会话 ID 传播器
	Store         Store      // 会话存储
	CtxSessionKey string     // 上下文中会话的键名
}

// GetSession 从上下文中获取会话，如果不存在则从请求中提取，并在存储中查找会话
func (m *Manager) GetSession(ctx *web.Context) (Session, error) {
	sess, ok := ctx.Get(m.CtxSessionKey)
	if ok {
		return sess.(Session), nil
	}

	sessId, err := m.Propagator.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}
	res, err := m.Store.Get(ctx.Req.Context(), sessId)
	if err != nil {
		return nil, err
	}
	ctx.Set(m.CtxSessionKey, res)
	return res, nil
}

// InitSession 初始化会话，并将其存储到存储中，然后将会话 ID 注入到 HTTP 响应中
func (m *Manager) InitSession(ctx *web.Context, sessId string) (Session, error) {
	existId, err := m.Propagator.Extract(ctx.Req)
	if err == nil {
		_ = m.Store.Remove(ctx.Req.Context(), existId)
	}
	sess, err := m.Store.Generate(ctx.Req.Context(), sessId)
	if err != nil {
		return nil, err
	}
	ctx.Set(m.CtxSessionKey, sess)
	// 注入http response
	err = m.Propagator.Inject(sess.ID(), ctx.Resp)
	return sess, err
}

// RemoveSession 移除会话，包括从存储中删除并清除上下文和响应中的会话信息
func (m *Manager) RemoveSession(ctx *web.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	err = m.Store.Remove(ctx.Req.Context(), sess.ID())
	if err != nil {
		return err
	}
	ctx.Set(m.CtxSessionKey, nil)
	// response header中清除session信息
	return m.Propagator.Clean(ctx.Resp)
}

// RefreshSession 刷新会话，更新会话的过期时间。
func (m *Manager) RefreshSession(ctx *web.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	return m.Store.Refresh(ctx.Req.Context(), sess.ID())
}

// SaveSession 将会话保存到存储中，并更新上下文中的会话信息。
func (m *Manager) SaveSession(ctx *web.Context, sess Session) error {
	err := m.Store.Set(ctx.Req.Context(), sess)
	if err != nil {
		return err
	}
	ctx.Set(m.CtxSessionKey, sess)
	return nil
}
