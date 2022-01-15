package server

import (
	"context"
)

func dbGetUser(ctx context.Context, uid int64) (*User, error) {
	userMap, err := dbBatchGetUser(ctx, []int64{uid})
	if err != nil {
		return nil, err
	}
	return userMap[uid], nil
}

func dbBatchGetUser(ctx context.Context, uids []int64) (map[int64]*User, error) {
	users := []*User{}
	err := slaveCli.Model(&User{}).Where("uid in (?)", uids).Find(&users).Error
	if err != nil {
		excLog.Printf("ctx %v dbBatchGetUser uids %v err %v", ctx, uids, err)
		return nil, err
	}
	userMap := make(map[int64]*User, len(uids))
	for _, v := range users {
		userMap[v.UID] = v
	}
	return userMap, nil
}

func dbAddUser(ctx context.Context, user *User) error {
	err := dbCli.Create(user).Error
	if err != nil {
		excLog.Printf("ctx %v dbAddUser user %v err %v", ctx, user, err)
	}
	return err
}

func dbAddBlack(ctx context.Context, uid, targetID int64, blackType int32) error {
	userBlack := &UserBlack{
		UID:           uid,
		BlackTargetID: targetID,
		BlackType:     BlackTypeArticle,
	}
	err := dbCli.Create(&userBlack).Error
	if err != nil {
		excLog.Printf("ctx %v dbAddBlack uid %v target_id %v black_type %v err %v", ctx, uid, targetID, blackType, err)
	}
	return err
}

func dbDelBlack(ctx context.Context, uid, targetID int64, blackType int32) error {
	err := dbCli.Where("uid = ? and target_id = ? and black_type = ?", uid, targetID, blackType).Delete(&UserBlack{}).Error
	if err != nil {
		excLog.Printf("ctx %v dbDelBlack uid %v target_id %v black_type %v err %v", ctx, uid, targetID, blackType, err)
	}
	return err
}

func dbAddCollection(ctx context.Context, uid, targetID int64) error {
	coll := &Collection{
		UID:      uid,
		TargetID: targetID,
	}
	err := dbCli.Create(coll).Error
	if err != nil {
		excLog.Printf("ctx %v dbAddCollection uid %v target_id %v err %v", ctx, uid, targetID, err)
	}
	return err
}

func dbDelCollection(ctx context.Context, uid, targetID int64) error {
	err := dbCli.Where("uid = ? and target_id = ?", uid, targetID).Delete(&Collection{}).Error
	if err != nil {
		excLog.Printf("ctx %v dbDelCollection uid %v target_id %v err %v", ctx, uid, targetID, err)
	}
	return err
}

func dbAddBrowse(ctx context.Context, uid, toUID int64) error {
	history := &BrowseHistory{
		UID:   uid,
		ToUID: toUID,
	}
	err := dbCli.Create(&history).Error
	if err != nil {
		excLog.Printf("ctx %v dbAddBrowse uid %v to_uid %v err %v", ctx, uid, toUID, err)
	}
	return err
}
