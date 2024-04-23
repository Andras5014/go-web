package session

import "context"

// Store 会话存储接口
type Store interface {
	Get(ctx context.Context, id string) (Session, error)
	Set(ctx context.Context, sess Session) error
	Generate(ctx context.Context, id string) (Session, error)
	Remove(ctx context.Context, id string) error
	Refresh(ctx context.Context, id string) error
}

// StoreOption 用于设置存储选项的函数类型
type StoreOption func(store Store)

// WithSessionBuilder 用于设置存储中会话生成器的选项
func WithSessionBuilder(builder Builder) StoreOption {
	return func(store Store) {
		switch s := store.(type) {
		case *MemoStore:
			s.sessionBuilder = builder
		case *RedisStore:
			s.sessionBuilder = builder
		}
	}
}

// WithSerializer 用于设置存储中序列化器的选项
func WithSerializer(serializer Serializer) StoreOption {
	return func(store Store) {
		switch s := store.(type) {
		case *RedisStore:
			s.serializer = serializer
		case *MemoStore:
			panic("memo Store needn't serializer")
		}
	}
}
