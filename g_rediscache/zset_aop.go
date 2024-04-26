package g_rediscache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"time"
)

// ZsetAop 中，EmptyExpires 使用无效
type ZSetOptions struct {
	Options

	// 如果为true，表示返回map，
	// key是member，value是score
	// NOTICE!!! 并且返回结果的map的value一定是float64类型
	IsMap bool

	Desc    bool
	ByScore bool

	// 如果是struct类型，需要指定一下，使用哪个字段作为score，否则score默认是0
	ScoreField string

	// 如果是ByScore，必须需要指定Min和Max
	Min    int
	Max    int
	Offset int64
	Count  int64

	// 如果不是ByScore, 需要指定Start, Stop
	Start int64
	Stop  int64
}

// NOTICE!!! 如果fallback返回结果的map的value一定是float64类型
func ZSetAop(options *ZSetOptions, fallback func() (interface{}, error)) (interface{}, bool, error) {
	zrangeBy := redis.ZRangeBy{
		Min:    strconv.Itoa(options.Min),
		Max:    strconv.Itoa(options.Max),
		Offset: options.Offset,
		Count:  options.Count,
	}
	if options.Stop == 0 {
		options.Stop = -1
	}
	res, err := getFromCache(options, zrangeBy)
	// 从缓存读取数据错误 直接返回
	if err != nil {
		return nil, false, err
	}
	// 返回值非空 且无报错
	if res != nil && err == nil {
		return res, true, nil
	}
	// 返回值为空 且无报错 则缓存中无数据（且不为空标记） 需要reload 执行fallback
	logrus.Info("[REDIS][ZSET] cant get value from redis cache, maybe load from db!")

	fResult, err := fallback()
	if err != nil {
		return nil, false, err
	}
	if fResult == nil {
		return nil, false, nil
	}
	rewriteCount := 0
	var members []*redis.Z
	if options.IsMap {

		for k, score := range fResult.(map[interface{}]float64) {
			cacheV, isEmpty, err := GetCacheValueItem(k)
			if err != nil {
				logrus.Warn("[REDIS][ZSET] GetCacheValueItem error!", err)
				continue
			}
			if !isEmpty {
				members = append(members, &redis.Z{Member: cacheV, Score: score})
				rewriteCount++
			}
		}
	} else {
		for _, resultItem := range fResult.([]interface{}) {
			var score float64 = 0
			// 看一下struct里面的作为score的field是否有正确的值
			if options.ScoreField != "" {
				iv := reflect.ValueOf(&resultItem)
				if reflect.TypeOf(resultItem).Kind() == reflect.Struct {
					ivf := iv.Elem().Elem().FieldByName(options.ScoreField)
					if ivf.IsValid() {
						vv, success := Number2Float64(ivf.Interface(), ivf.Kind())
						if success {
							score = vv
						}
					}
				}
			}
			cacheV, isEmpty, err := GetCacheValueItem(resultItem)
			if err != nil {
				logrus.Warn("[REDIS][ZSET] GetCacheValueItem error!", err)
				continue
			}
			if !isEmpty {
				members = append(members, &redis.Z{Member: cacheV, Score: score})
				rewriteCount++
			}
		}

	}
	if len(members) > 0 {
		if err := GetRedisClient().ZAdd(options.Ctx, options.Key, members...).Err(); err != nil {
			return nil, false, err
		}
	}
	if rewriteCount > 0 {
		if err := GetRedisClient().Expire(options.Ctx, options.Key, options.Expires).Err(); err != nil {
			return nil, false, err
		}
	} else {
		// 空值回填
		if rewriteCount == 0 && options.EmptyExpires > 0 {
			if err := GetRedisClient().ZAdd(options.Ctx, options.Key, &redis.Z{Member: EmptyFlag}).Err(); err != nil {
				return nil, false, err
			}
			GetRedisClient().Expire(options.Ctx, options.Key, options.EmptyExpires).Val()
			logrus.Warn("[REDIS][ZSET] cache empty value, key:", options.Key)
		}
	}
	// 回填完成 再次从缓存中取排序好的数据
	res2, err := getFromCache(options, zrangeBy)
	return res2, false, err
}

func getFromCache(options *ZSetOptions, zrangeBy redis.ZRangeBy) (interface{}, error) {
	var cacheVs []redis.Z
	var err error = nil
	client := GetRedisClient()
	if options.Desc && options.ByScore {
		cacheVs, err = client.ZRevRangeByScoreWithScores(options.Ctx, options.Key, &zrangeBy).Result()
	} else if options.Desc {
		cacheVs, err = client.ZRevRangeWithScores(options.Ctx, options.Key, options.Start, options.Stop).Result()
	} else if options.ByScore {
		cacheVs, err = client.ZRangeByScoreWithScores(options.Ctx, options.Key, &zrangeBy).Result()
	} else {
		cacheVs, err = client.ZRangeWithScores(options.Ctx, options.Key, options.Start, options.Stop).Result()
	}
	if err != nil {
		return nil, err
	}
	var result []interface{}
	var mapResult = make(map[interface{}]float64)
	if len(cacheVs) > 0 {
		if len(cacheVs) == 1 && cacheVs[0].Member == EmptyFlag {
			if options.IsMap {
				return mapResult, nil
			} else {
				return result, nil
			}
		}
		for _, cacheV := range cacheVs {
			rtv := reflect.New(options.Rt)
			rv := rtv.Interface()
			err := json.Unmarshal([]byte(cacheV.Member.(string)), rv)
			if err != nil {
				return nil, err
			}
			vv := reflect.ValueOf(rv).Elem().Interface()

			if options.IsMap {
				mapResult[vv] = cacheV.Score
			} else {
				result = append(result, vv)
			}
		}
		if options.IsMap {
			return mapResult, nil
		} else {
			return result, nil
		}
	}
	return nil, nil
}

type ZSetAopProxy struct {
	options ZSetOptions
}

func (p *ZSetAopProxy) WithExpires(expires time.Duration) *ZSetAopProxy {
	p.options.Expires = expires
	return p
}

func (p *ZSetAopProxy) WithEmptyExpires(emptyExpires time.Duration) *ZSetAopProxy {
	p.options.EmptyExpires = emptyExpires
	return p
}

func (p *ZSetAopProxy) WithIsMap(isMap bool) *ZSetAopProxy {
	p.options.IsMap = isMap
	return p
}

func (p *ZSetAopProxy) WithDesc(desc bool) *ZSetAopProxy {
	p.options.Desc = desc
	return p
}

func (p *ZSetAopProxy) WithByScore(byScore bool) *ZSetAopProxy {
	p.options.ByScore = byScore
	return p
}

func (p *ZSetAopProxy) WithScoreField(scoreField string) *ZSetAopProxy {
	p.options.ScoreField = scoreField
	return p
}

func (p *ZSetAopProxy) WithMin(min int) *ZSetAopProxy {
	p.options.Min = min
	return p
}

func (p *ZSetAopProxy) WithMax(max int) *ZSetAopProxy {
	p.options.Max = max
	return p
}

func (p *ZSetAopProxy) WithOffset(offset int64) *ZSetAopProxy {
	p.options.Offset = offset
	return p
}

func (p *ZSetAopProxy) WithCount(count int64) *ZSetAopProxy {
	p.options.Count = count
	return p
}
func (p *ZSetAopProxy) WithStart(start int64) *ZSetAopProxy {
	p.options.Start = start
	return p
}

func (p *ZSetAopProxy) WithStop(stop int64) *ZSetAopProxy {
	p.options.Stop = stop
	return p
}

func (p *ZSetAopProxy) Then(f func() (interface{}, error)) (interface{}, bool, error) {
	return ZSetAop(&p.options, f)
}

func UseZSetAop(ctx context.Context, key string, rt reflect.Type) *ZSetAopProxy {
	return &ZSetAopProxy{ZSetOptions{Options: Options{Ctx: ctx, Key: key, Rt: rt}}}
}
