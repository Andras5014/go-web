package session

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/redis/go-redis/v9"
	"reflect"
	"time"
)

type RedisStore struct {
	cmd            redis.Cmdable
	exp            time.Duration
	sessionBuilder Builder
	serializer     Serializer
}

// Serializer 序列化Session接口
type Serializer interface {
	RegisterType(sess Session)      // 注册会话类型
	Encode(Session) ([]byte, error) // 编码会话数据
	Decode([]byte) (Session, error) // 解码会话数据
}

type DefaultSerializer struct {
	sessType reflect.Type
}

// RegisterType 注册会话类型
func (d *DefaultSerializer) RegisterType(sess Session) {
	d.sessType = reflect.TypeOf(sess).Elem()
}

// Encode 编码会话数据
func (d *DefaultSerializer) Encode(sess Session) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(sess)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode 解码会话数据
func (d *DefaultSerializer) Decode(data []byte) (Session, error) {
	buf := bytes.NewBuffer(data)
	res := reflect.New(d.sessType)
	err := gob.NewDecoder(buf).DecodeValue(res)
	if err != nil {
		return nil, err
	}
	return res.Interface().(Session), nil
}

// NewRedisStore 创建一个新的 Redis 存储
func NewRedisStore(cmd redis.Cmdable, expiration time.Duration, opts ...StoreOption) *RedisStore {
	res := &RedisStore{
		exp:            expiration,
		cmd:            cmd,
		sessionBuilder: DefaultBuilder,
		serializer:     &DefaultSerializer{},
	}
	for _, opt := range opts {
		opt(res)
	}
	res.serializer.RegisterType(res.sessionBuilder(res, ""))

	return res
}

// Get 从 Redis 获取会话数据
func (r *RedisStore) Get(ctx context.Context, id string) (Session, error) {
	val, err := r.cmd.Get(ctx, id).Bytes()
	if err != nil {
		return nil, err
	}
	return r.serializer.Decode(val)
}

// Set 将会话数据存储到 Redis
func (r *RedisStore) Set(ctx context.Context, sess Session) error {
	data, err := r.serializer.Encode(sess)
	if err != nil {
		return err
	}
	return r.cmd.Set(ctx, sess.ID(), data, r.exp).Err()
}

// Generate 创建一个新的会话并存储到 Redis
func (r *RedisStore) Generate(ctx context.Context, id string) (Session, error) {
	// 使用指定的会话生成器创建新的会话实例
	sess := r.sessionBuilder(r, id)
	err := r.Set(ctx, sess)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// Remove 从 Redis 中删除会话数据
func (r *RedisStore) Remove(ctx context.Context, id string) error {
	return r.cmd.Del(ctx, id).Err()
}

// Refresh 更新存储中指定 ID 的会话的刷新时间
func (r *RedisStore) Refresh(ctx context.Context, id string) error {
	return r.cmd.Expire(ctx, id, r.exp).Err()
}
