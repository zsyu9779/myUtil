package g_rediscache

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type SimpleOptions struct {
	Options
}

func SimpleAop(options *SimpleOptions, fallback func() (interface{}, error)) (interface{}, bool, error) {
	err := options.validate()
	if err != nil {
		return nil, false, err
	}
	cacheV, err := GetRedisClient().Get(options.Ctx, options.Key).Result()
	if cacheV != "" {
		rtv := reflect.New(options.Rt)
		rv := rtv.Interface()
		if cacheV == EmptyFlag {
			return nil, true, nil
		}
		if options.Rt.Kind() == reflect.String {
			return cacheV, true, nil
		}
		err := json.Unmarshal([]byte(cacheV), rv)
		return reflect.ValueOf(rv).Elem().Interface(), true, err
	}
	logrus.Warn("[REDIS][SIMPLE] cant get value from redis cache, maybe load from db!")
	var result interface{} = nil
	result, err = fallback()
	if err != nil {
		return nil, false, err
	}
	// 是否回填cache成功
	rewriteSuccess := false
	if result != nil {
		cacheV, isEmpty, err := GetCacheValueItem(result)
		if err != nil {
			return nil, false, err
		}
		if !isEmpty {
			if options.Expires == 0 {
				options.Expires = defaultExpire
			}
			GetRedisClient().Set(options.Ctx, options.Key, cacheV, options.Expires).Val()
			rewriteSuccess = true
		}
	}
	// 是否需要存储空值
	if !rewriteSuccess && options.EmptyExpires > 0 {
		GetRedisClient().Set(options.Ctx, options.Key, EmptyFlag, options.EmptyExpires).Val()
		logrus.Warn("[REDIS][SIMPLE] cache empty value, key:", options.Key)
	}
	return result, false, nil
}

type SimpleAopProxy struct {
	options SimpleOptions
}

func (p *SimpleAopProxy) WithExpires(expires time.Duration) *SimpleAopProxy {
	p.options.Expires = expires
	return p
}

func (p *SimpleAopProxy) WithEmptyExpires(emptyExpires time.Duration) *SimpleAopProxy {
	p.options.EmptyExpires = emptyExpires
	return p
}

func (p *SimpleAopProxy) Then(f func() (interface{}, error)) (interface{}, bool, error) {
	return SimpleAop(&p.options, f)
}

func UseSimpleAop(ctx context.Context, key string, rt reflect.Type) *SimpleAopProxy {
	return &SimpleAopProxy{SimpleOptions{Options{Ctx: ctx, Key: key, Rt: rt}}}
}
