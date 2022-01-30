package server

import (
	"context"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"userservice/conf"
	"userservice/rpc/user/pb"
)

type UserService struct {
}

var (
	mcCli    *memcache.Client
	redisCli redis.Cmdable
	dbCli    *gorm.DB
	slaveCli *gorm.DB
)

func InitService(config *conf.Conf) error {
	var err error
	mcCli = conf.GetMC(config.MC.Addr)
	redisCli = conf.GetRedisCluster(config.RedisCluster.Addr)
	dbCli, err = conf.GetGorm(fmt.Sprintf(conf.MysqlAddr, config.Mysql.User, config.Mysql.Password, config.Mysql.Host, config.Mysql.Port, config.Mysql.DB))
	if err != nil {
		return err
	}
	slaveCli, err = conf.GetGorm(fmt.Sprintf(conf.MysqlAddr, config.Slave.User, config.Slave.Password, config.Slave.Host, config.Slave.Port, config.Slave.DB))
	return err
}

func (us *UserService) GetUserinfo(ctx context.Context, req *user_service.UserInfoRequest, res *user_service.UserInfoResponse) error {
	userinfo, err := getUserinfo(ctx, req.Uid)
	if err != nil {
		return err
	}
	res.Userinfo = userinfo
	return nil
}

func (us *UserService) GetBatchUserinfo(ctx context.Context, req *user_service.UserInfoBatchRequest, res *user_service.UserInfoBatchResponse) error {
	userinfoMap, err := getBatchUserinfo(ctx, req.Uids)
	if err != nil {
		return err
	}
	res.Userinfos = userinfoMap
	return nil
}

func (us *UserService) GetHistoryBrowse(ctx context.Context, req *user_service.FeedListRequest, res *user_service.FeedListResponse) error {
	targetIDs, hasMore, err := getHistoryBrowse(ctx, req.Uid, req.LastID, req.Offset)
	if err != nil {
		return err
	}
	res.TargetIds = targetIDs
	res.HasMore = hasMore
	return nil
}

func (us *UserService) GetBlackList(ctx context.Context, req *user_service.BlackListRequest, res *user_service.FeedListResponse) error {
	targetIDs, hasMore, err := getBlackList(ctx, req.Uid, req.LastID, req.Offset, req.BlackType)
	if err != nil {
		return err
	}
	res.TargetIds = targetIDs
	res.HasMore = hasMore
	return nil
}

func (us *UserService) GetCollectionList(ctx context.Context, req *user_service.FeedListRequest, res *user_service.FeedListResponse) error {
	targetIDs, hasMore, err := getCollectionList(ctx, req.Uid, req.LastID, req.Offset)
	if err != nil {
		return err
	}
	res.TargetIds = targetIDs
	res.HasMore = hasMore
	return nil
}

func (us *UserService) Black(ctx context.Context, req *user_service.BlackRequest, res *user_service.EmptyResponse) error {
	err := black(ctx, req.BlackInfo.Uid, req.BlackInfo.TargetId, req.BlackInfo.BlackType)
	return err
}

func (us *UserService) CancelBlack(ctx context.Context, req *user_service.CancelBlackRequest, res *user_service.EmptyResponse) error {
	err := cancelBlack(ctx, req.BlackInfo.Uid, req.BlackInfo.TargetId, req.BlackInfo.BlackType)
	return err
}

func (us *UserService) Collection(ctx context.Context, req *user_service.CollectionRequest, res *user_service.EmptyResponse) error {
	err := collection(ctx, req.CollectionInfo.Uid, req.CollectionInfo.TargetId)
	return err
}

func (us *UserService) CancelCollection(ctx context.Context, req *user_service.CancelCollectionRequest, res *user_service.EmptyResponse) error {
	err := cancelCollection(ctx, req.CollectionInfo.Uid, req.CollectionInfo.TargetId)
	return err
}

func (us *UserService) AddBrowse(ctx context.Context, req *user_service.AddBrowseRequest, res *user_service.EmptyResponse) error {
	err := addBrowse(ctx, req.BrowseInfo.Uid, req.BrowseInfo.ToUid)
	return err
}

func (us *UserService) CreateUser(ctx context.Context, req *user_service.CreateUserRequest, res *user_service.EmptyResponse) error {
	err := createUser(ctx, req.Userinfo.Uid, req.Userinfo.Gender, req.Userinfo.Nickname, req.Userinfo.Introduction)
	return err
}
