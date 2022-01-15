package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis"
	"time"
	"userservice/global"
)

const (
	DefaultOffset = 10

	SleepTime = 500 * time.Millisecond

	BlackTypePerson  = 1
	BlackTypeArticle = 2

	MCKeyUserinfo    = "user_service_info_%v" // uid
	MCKeyUserinfoTTl = 5 * 60

	RedisKeyZBlackUser    = "user_service_black_user_%v"    // uid
	RedisKeyZBlackArticle = "user_service_black_article_%v" // article_id
	RedisKeyZCollection   = "user_service_collection_%v"    // uid
	RedisKeyZBrowse       = "user_service_browse_%v"        // uid
)

func cacheGetUser(ctx context.Context, uid int64) (*User, error) {
	userMap, missed, err := cacheBatchGetUser(ctx, []int64{uid})
	if err != nil || len(missed) != 0 {
		return nil, err
	}
	return userMap[uid], nil
}

func cacheBatchGetUser(ctx context.Context, uids []int64) (map[int64]*User, []int64, error) {
	keys := make([]string, 0, len(uids))
	for _, v := range uids {
		keys = append(keys, fmt.Sprintf(MCKeyUserinfo, v))
	}
	res, err := mcCli.GetMulti(keys)
	if err != nil {
		return nil, uids, err
	}
	userMap := make(map[int64]*User, len(uids))
	for _, v := range res {
		user := User{}
		err = json.Unmarshal(v.Value, &user)
		if err != nil {
			global.ExcLog.Printf("ctx %v cacheBatchGetUser user %v unmarshal err %v", ctx, v.Value, err)
			continue
		}
		userMap[user.UID] = &user
	}
	missed := make([]int64, 0, len(uids))
	for _, v := range uids {
		if _, ok := userMap[v]; !ok {
			missed = append(missed, v)
		}
	}
	return userMap, missed, nil
}

//func (dal *UserDAL) cacheGetHistoryBrowse(ctx context.Context, uid, cursor, offset int64) ([]int64, []int64, bool, error) {
//
//}
//
//func (dal *UserDAL) cacheSetHistoryBrowse(ctx context.Context, uid int64, articleIDs []int64) error {
//
//}

func cacheSetUser(ctx context.Context, user *User) error {
	buf, err := json.Marshal(user)
	if err != nil {
		return err
	}
	err = mcCli.Set(&memcache.Item{Key: fmt.Sprintf(MCKeyUserinfo, user.UID), Value: buf, Expiration: MCKeyUserinfoTTl})
	if err != nil {
		global.ExcLog.Printf("ctx %v cacheSetUser user %v err %v", ctx, user, err)
	}
	return err
}

func cacheBatchSetUser(ctx context.Context, userMap map[int64]*User) error {
	for k, v := range userMap {
		val, err := json.Marshal(v)
		if err != nil {
			global.ExcLog.Printf("ctx %v cacheBatchSetUser marshal user %v err %v", ctx, v, err)
			continue
		}
		err = mcCli.Set(&memcache.Item{Key: fmt.Sprintf(MCKeyUserinfo, k), Value: val, Expiration: MCKeyUserinfoTTl})
		if err != nil {
			global.ExcLog.Printf("ctx %v cacheBatchSetUser set mc user %v err %v", ctx, val, err)
			continue
		}
	}
	return nil
}

func cacheBlack(ctx context.Context, uid, targetID int64, blackType int32) error {
	switch blackType {
	case BlackTypePerson:
		return blackPerson(ctx, uid, targetID)
	case BlackTypeArticle:
		return blackArticle(ctx, uid, targetID)
	}
	global.ExcLog.Printf("ctx %v cacheBlack uid %v target_id %v black_type %v err", ctx, uid, targetID, blackType)
	return errors.New("black_type err")
}

func blackPerson(ctx context.Context, uid, toUID int64) error {
	key := fmt.Sprintf(RedisKeyZBlackUser, uid)
	sKey := fmt.Sprintf(RedisKeyZBlackUser, toUID)
	now := float64(time.Now().Unix())
	pipe := redisCli.Pipeline()
	pipe.ZAdd(key, redis.Z{Member: toUID, Score: now})
	pipe.ZAdd(sKey, redis.Z{Member: uid, Score: now})
	_, err := pipe.Exec()
	if err != nil {
		global.ExcLog.Printf("ctx %v blackPerson uid %v to_uid %v err %v", ctx, uid, toUID, err)
	}
	return err
}

func blackArticle(ctx context.Context, uid, articleID int64) error {
	key := fmt.Sprintf(RedisKeyZBlackArticle, uid)
	now := float64(time.Now().Unix())
	err := redisCli.ZAdd(key, redis.Z{Member: articleID, Score: now}).Err()
	if err != nil {
		global.ExcLog.Printf("ctx %v blackArticle uid %v article_id %v err %v", ctx, uid, articleID, err)
	}
	return err
}

func cacheCancelBlack(ctx context.Context, uid, targetID int64, blackType int32) error {
	switch blackType {
	case BlackTypePerson:
		return cancelBlackPerson(ctx, uid, targetID)
	case BlackTypeArticle:
		return cancelBlackArticle(ctx, uid, targetID)
	}
	global.ExcLog.Printf("ctx %v cacheCancelBlack uid %v target_id %v black_type %v err", ctx, uid, targetID, blackType)
	return errors.New("black_type err")
}

func cancelBlackPerson(ctx context.Context, uid, toUID int64) error {
	key := fmt.Sprintf(RedisKeyZBlackUser, uid)
	sKey := fmt.Sprintf(RedisKeyZBlackUser, toUID)
	pipe := redisCli.Pipeline()
	pipe.ZRem(key, toUID)
	pipe.ZRem(sKey, uid)
	_, err := pipe.Exec()
	if err != nil {
		global.ExcLog.Printf("ctx %v cancelBlackPerson uid %v target_id %v err %v", ctx, uid, toUID, err)
	}
	return err
}

func cancelBlackArticle(ctx context.Context, uid, articleID int64) error {
	key := fmt.Sprintf(RedisKeyZBlackArticle, uid)
	err := redisCli.ZRem(key, articleID).Err()
	if err != nil {
		global.ExcLog.Printf("ctx %v cancelBlackArticle uid %v article_id %v err %v", ctx, uid, articleID, err)
	}
	return err
}

func cacheDelCollection(ctx context.Context, uid int64) error {
	key := fmt.Sprintf(RedisKeyZCollection, uid)
	err := redisCli.Del(key).Err()
	if err != nil {
		global.ExcLog.Printf("ctx %v cacheDelCollection uid %v err %v", ctx, uid, err)
	}
	return err
}

func cacheDelBrowse(ctx context.Context, uid int64) error {
	key := fmt.Sprintf(RedisKeyZBrowse, uid)
	err := redisCli.Del(key).Err()
	if err != nil {
		global.ExcLog.Printf("ctx %v cacheDelBrowse uid %v err %v", ctx, uid, err)
	}
	return err
}
