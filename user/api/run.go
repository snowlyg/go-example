package api

import (
	"github.com/micro/cli"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
	"github.com/micro/go-plugins/wrapper/ratelimiter/uber"
	ot "github.com/micro/go-plugins/wrapper/trace/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/xiaomeng79/go-example/cinit"
	"github.com/xiaomeng79/go-example/internal/mywrapper"
	pb "github.com/xiaomeng79/go-example/user/srv/proto"
	"github.com/xiaomeng79/go-log"
	"time"
)

//定义services名称
const SN = "com.example.api.user"

//运行
func Run() {
	//初始化
	cinit.InitOption(SN, "trace")
	//新建服务
	serviceName := cinit.Config.Service.Name
	serviceVersion := cinit.Config.Service.Version
	service := micro.NewService(
		micro.Name(serviceName),
		micro.Version(serviceVersion),
		micro.RegisterTTL(time.Second*30),
		micro.RegisterInterval(time.Second*15),
		micro.WrapClient(ot.NewClientWrapper(opentracing.GlobalTracer())),
		micro.WrapHandler(ot.NewHandlerWrapper(opentracing.GlobalTracer())),
		micro.WrapClient(mywrapper.LogClientWrap),
		micro.WrapHandler(mywrapper.LogServerWrap),
		micro.WrapHandler(ratelimit.NewHandlerWrapper(cinit.Config.Service.RateTime)),
	)
	// 优雅关闭
	service.Server().Init(
		server.Wait(false),
	)
	log.Info("创建服务:名称:" + serviceName + ",版本:" + serviceVersion)
	//初始化配置
	service.Init(
		micro.Action(func(c *cli.Context) {
			//注册服务
			//pb.RegisterAccountHandler(service.Server(),handler.NewAccountService(), server.InternalHandler(true))
			//pb.RegisterUserHandler(service.Server(), handler.NewUserService(), server.InternalHandler(true))
		}),
		micro.AfterStop(func() error {
			log.Info("停止服务")
			//停止配置
			cinit.Close()
			return nil
		}),
		micro.AfterStart(func() error {
			return nil
		}),
	)
	//新建一个句柄
	service.Server().Handle(
		service.Server().NewHandler(
			&User{Client: pb.NewUserService("com.example.srv.user", service.Client())},//默认的客户端参数
			//设置客户端参数
			//&User{Client: pb.NewUserService("com.example.srv.user", client.NewClient(
			//	client.Retries(2),//重试次数
			//	client.WrapCall(ot.NewCallWrapper(opentracing.GlobalTracer())),//客户端包装
			//	client.DialTimeout(time.Second*30),//客户端延迟断开
			//	))},
		),
	)
	//启动service
	if err := service.Run(); err != nil {
		log.Fatal("启动服务失败" + err.Error())
	}
}
