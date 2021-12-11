package main

import (
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/registry/etcd"
	"userservice/conf"
	"userservice/rpc/user/pb"
	"userservice/server"
)

var (
	userConf *conf.Conf
	err      error
)

func main() {
	userConf, err = conf.LoadYaml(conf.UserConfPath)
	if err != nil {
		panic(err)
	}

	err = server.InitService(userConf)
	if err != nil {
		panic(err)
	}

	etcdRegistry := etcd.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{userConf.Etcd.Addr}
	})

	service := micro.NewService(
		micro.Name(userConf.Grpc.Name),
		micro.Address(userConf.Grpc.Addr),
		micro.Registry(etcdRegistry),
	)
	service.Init()
	err = user_service.RegisterUserServerHandler(
		service.Server(),
		new(server.UserService),
	)
	if err != nil {
		panic(err)
	}
	err = service.Run()
	if err != nil {
		panic(err)
	}
}
