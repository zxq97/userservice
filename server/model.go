package server

import "userservice/rpc/user/pb"

type User struct {
	UID          int64  `json:"uid"`
	Nickname     string `json:"nickname"`
	Introduction string `json:"introduction"`
	Gender       int32  `json:"gender"`
}

type UserBlack struct {
	UID           int64 `json:"uid"`
	BlackTargetID int64 `json:"black_target_id"`
	BlackType     int   `json:"black_type"`
}

type Collection struct {
	UID      int64 `json:"uid"`
	TargetID int64 `json:"target_id"`
}

type BrowseHistory struct {
	UID   int64 `json:"uid"`
	ToUID int64 `json:"to_uid"`
}

func (u *User) toUserinfo() *user_service.UserInfo {
	return &user_service.UserInfo{
		Uid:          u.UID,
		Nickname:     u.Nickname,
		Introduction: u.Introduction,
		Gender:       u.Gender,
	}
}

func (t *UserBlack) TableName() string {
	return "user_black"
}

func (t *Collection) TableName() string {
	return "collection"
}

func (t *BrowseHistory) TableName() string {
	return "browse_history"
}
