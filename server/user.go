package server

import (
	"context"
	"github.com/go-redis/redis"
	"userservice/global"
	"userservice/rpc/user/pb"
	"userservice/util/concurrent"
)

func getUserinfo(ctx context.Context, uid int64) (*user_service.UserInfo, error) {
	user, err := cacheGetUser(ctx, uid)
	if err != nil || user == nil {
		user, err = dbGetUser(ctx, uid)
		if err != nil || user == nil {
			return nil, err
		}
		concurrent.Go(func() {
			_ = cacheSetUser(ctx, user)
		})
	}
	return user.toUserinfo(), nil
}

func getBatchUserinfo(ctx context.Context, uids []int64) (map[int64]*user_service.UserInfo, error) {
	userMap := make(map[int64]*user_service.UserInfo, len(uids))
	cacheUserMap, missed, err := cacheBatchGetUser(ctx, uids)
	for k, v := range cacheUserMap {
		userMap[k] = v.toUserinfo()
	}
	if err != nil || len(missed) != 0 {
		var dbUserMap map[int64]*User
		dbUserMap, err = dbBatchGetUser(ctx, uids)
		if err != nil {
			return userMap, nil
		}
		concurrent.Go(func() {
			_ = cacheBatchSetUser(ctx, dbUserMap)
		})
		for k, v := range dbUserMap {
			userMap[k] = v.toUserinfo()
		}
	}
	return userMap, nil
}

func getHistoryBrowse(ctx context.Context, uid, lastID, offset int64) ([]int64, bool, error) {
	if offset == 0 {
		offset = DefaultOffset
	}
	return nil, false, nil
}

func getBlackList(ctx context.Context, uid, lastID, offset int64, blackType int32) ([]int64, bool, error) {
	if offset == 0 {
		offset = DefaultOffset
	}
	return nil, false, nil
}

func getCollectionList(ctx context.Context, uid, lastID, offset int64) ([]int64, bool, error) {
	if offset == 0 {
		offset = DefaultOffset
	}
	return nil, false, nil
}

func black(ctx context.Context, uid, targetID int64, blackType int32) error {
	err := dbAddBlack(ctx, uid, targetID, blackType)
	if err != nil {
		return err
	}
	concurrent.Go(func() {
		_ = cacheBlack(ctx, uid, targetID, blackType)
	})
	return nil
}

func cancelBlack(ctx context.Context, uid, targetID int64, blackType int32) error {
	err := dbDelBlack(ctx, uid, targetID, blackType)
	if err != nil {
		return err
	}
	return cacheCancelBlack(ctx, uid, targetID, blackType)
}

func collection(ctx context.Context, uid, targetID int64) error {
	err := dbAddCollection(ctx, uid, targetID)
	if err != nil {
		return err
	}
	return cacheDelCollection(ctx, uid)
}

func cancelCollection(ctx context.Context, uid, targetID int64) error {
	err := dbDelCollection(ctx, uid, targetID)
	if err != nil {
		return err
	}
	return cacheDelCollection(ctx, uid)
}

func addBrowse(ctx context.Context, uid, toUID int64) error {
	err := dbAddBrowse(ctx, uid, toUID)
	if err != nil {
		return err
	}
	return cacheDelBrowse(ctx, uid)
}

func createUser(ctx context.Context, uid int64, gender int32, nickname, introduction string) error {
	user := &User{
		UID:          uid,
		Nickname:     nickname,
		Gender:       gender,
		Introduction: introduction,
	}
	err := dbAddUser(ctx, user)
	if err != nil {
		return err
	}
	concurrent.Go(func() {
		_ = cacheSetUser(ctx, user)
	})
	return nil
}

func getRankByID(ctx context.Context, lastID, key string) (int64, error) {
	rank, err := redisCli.ZRank(key, lastID).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		global.ExcLog.Printf("ctx %v getRankByID key %v last_id %v err %v", ctx, key, lastID, err)
		return 0, err
	}
	return rank, nil
}
