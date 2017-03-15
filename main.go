package main

import (
	"fmt"
	"time"

	"github.com/1046102779/official_account/conf"
	"github.com/1046102779/official_account/models"
	_ "github.com/1046102779/official_account/routers"

	"github.com/astaxie/beego"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx"
	"github.com/smallnest/rpcx/codec"
	"github.com/smallnest/rpcx/plugin"
)

func startRPCService(rpcAddr string, etcdAddr string, officialAccountServer *models.OfficialAccountServer) {
	server := rpcx.NewServer()
	rplugin := &plugin.EtcdRegisterPlugin{
		ServiceAddress: "tcp@" + rpcAddr,
		EtcdServers:    []string{etcdAddr},
		BasePath:       fmt.Sprintf("/%s/%s", beego.BConfig.RunMode, "rpcx"),
		Metrics:        metrics.NewRegistry(),
		Services:       make([]string, 0),
		UpdateInterval: time.Minute,
	}
	rplugin.Start()
	server.PluginContainer.Add(rplugin)
	server.PluginContainer.Add(plugin.NewMetricsPlugin())
	server.RegisterName("official_accounts", officialAccountServer, "weight=1&m=devops")
	server.ServerCodecFunc = codec.NewProtobufServerCodec
	server.Serve("tcp", rpcAddr)
}
func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	fmt.Println("main starting...")
	go startRPCService(conf.RpcAddr, conf.EtcdAddr, &models.OfficialAccountServer{})

	beego.Run()
}
