package session

import (
	"context"

	"github.com/fanjindong/go-cache"
	"time"
)

type MemoStore struct {
	store          cache.ICache  // 存储会话数据的缓存
	exp            time.Duration // 会话过期时间
	sessionBuilder Builder       // 会话生成器
}

// NewMemoStore 创建一个新的内存存储
func NewMemoStore(expiration time.Duration, opts ...StoreOption) *MemoStore {
	res := &MemoStore{
		store:          cache.NewMemCache(),
		exp:            expiration,
		sessionBuilder: DefaultBuilder,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

// Get 从内存缓存中获取会话数据
func (m *MemoStore) Get(ctx context.Context, id string) (Session, error) {
	sess, ok := m.store.Get(id)
	if !ok {
		return nil, ErrKeyNotFound
	}
	// 检查获取到的数据是否是 Session 类型
	if s, ok := sess.(Session); ok {
		return s, nil
	} else {
		return nil, ErrKeyNotFound
	}
}

// Set 将会话数据存储到内存缓存中
func (m *MemoStore) Set(ctx context.Context, sess Session) error {
	// 将会话数据存储到缓存中，并设置过期时间
	ok := m.store.Set(sess.ID(), sess, cache.WithEx(m.exp))
	if !ok {
		return ErrSaveFailed
	}
	return nil
}

// Generate 创建一个新的会话并存储到内存缓存中
func (m *MemoStore) Generate(ctx context.Context, id string) (Session, error) {
	s := m.sessionBuilder(m, id)
	ok := m.store.Set(id, s, cache.WithEx(m.exp))
	if !ok {
		return nil, ErrSaveFailed
	}
	return s, nil
}

// Remove 从内存缓存中删除会话数据
func (m *MemoStore) Remove(ctx context.Context, id string) error {
	m.store.Del(id)
	return nil
}

// Refresh 更新内存缓存中的会话刷新时间
func (m *MemoStore) Refresh(ctx context.Context, id string) error {
	ok := m.store.Expire(id, m.exp)
	if !ok {
		return ErrKeyNotFound
	}
	return nil
}
